package jsonex

import (
	"io"
)

// parseState represents the current state of the JSON parser (unexported)
type parseState int

const (
	stateValue parseState = iota
	stateObjectStart
	stateObjectKey
	stateObjectColon
	stateObjectValue
	stateObjectComma
	stateArrayStart
	stateArrayValue
	stateArrayComma
	stateEnd
)

// parser handles JSON syntax parsing and validation (unexported)
type parser struct {
	scanner *scanner
	options options
	depth   int
	state   parseState
}

// newParser creates a new parser
func newParser(reader io.Reader, opts options) *parser {
	return &parser{
		scanner: newScanner(reader, opts.bufferSize),
		options: opts,
		depth:   0,
		state:   stateValue,
	}
}

// parseNext extracts the next complete JSON object or array from the stream
// This is used by the Decoder for streaming processing
func (p *parser) parseNext() ([]byte, error) {
	// Find the start of JSON (object or array)
	startByte, err := p.scanner.findJSONStart()
	if err != nil {
		return nil, err
	}

	// Reset parser state
	p.depth = 0
	p.state = stateValue

	// Create buffer to collect the JSON
	buf := getBuffer()
	defer putBuffer(buf)

	// Start parsing from the found position
	return p.parseValue(startByte, buf)
}

// parseLongest finds and extracts the longest valid JSON from byte data
// This is used by the Unmarshal function for batch processing
func parseLongest(data []byte, opts options) ([]byte, error) {
	var longestJSON []byte
	var bestLength int
	var hasCustomOptions = opts.maxDepth != 1000 || opts.bufferSize != 4096

	// Try parsing from each potential JSON start position
	for i := 0; i < len(data); i++ {
		if data[i] == '{' || data[i] == '[' {
			// Try to parse JSON starting from this position
			jsonData, length, err := tryParseFromPosition(data[i:], opts)
			if err == nil && length > bestLength {
				longestJSON = make([]byte, length)
				copy(longestJSON, jsonData)
				bestLength = length
			} else if err != nil {
				// If we have custom options (especially depth limits) and encounter depth errors,
				// return the error immediately to enforce limits strictly
				if hasCustomOptions && isDepthError(err) {
					return nil, err
				}
			}
		}
	}

	// If we found valid JSON, return it
	if longestJSON != nil {
		return longestJSON, nil
	}

	return nil, newInvalidJSONError(position{}, "no valid JSON found")
}

// isDepthError checks if an error is related to depth limits
func isDepthError(err error) bool {
	if jsonErr, ok := err.(*Error); ok {
		return jsonErr.Type == ErrSyntax &&
			(jsonErr.Message == "maximum nesting depth exceeded")
	}
	return false
}

// tryParseFromPosition attempts to parse JSON from a specific position
func tryParseFromPosition(data []byte, opts options) ([]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, newEOFError(position{}, "empty data")
	}

	// Create a temporary scanner for this data
	reader := &bytesReader{data: data, pos: 0}
	parser := newParser(reader, opts)

	// Try to parse
	result, err := parser.parseNext()
	if err != nil {
		return nil, 0, err
	}

	return result, len(result), nil
}

// bytesReader implements io.Reader for byte slices
type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.pos:])
	r.pos += n

	if r.pos >= len(r.data) {
		return n, io.EOF
	}

	return n, nil
}

// parseValue parses a JSON value (object or array)
func (p *parser) parseValue(startByte byte, buf *buffer) ([]byte, error) {
	switch startByte {
	case '{':
		return p.parseObject(buf)
	case '[':
		return p.parseArray(buf)
	default:
		return nil, newSyntaxError(p.scanner.position(), "expected '{' or '['")
	}
}

// parseObject parses a JSON object
func (p *parser) parseObject(buf *buffer) ([]byte, error) {
	p.depth++
	defer func() { p.depth-- }()

	if err := p.checkDepth(); err != nil {
		return nil, err
	}

	buf.writeByte('{')

	// Consume the opening brace
	b, err := p.scanner.next()
	if err != nil {
		return nil, err
	}
	if b != '{' {
		return nil, newSyntaxError(p.scanner.position(), "expected '{'")
	}

	// Skip whitespace
	if err := p.scanner.skipWhitespace(); err != nil {
		return nil, err
	}

	// Check for empty object
	if b, err := p.scanner.peek(); err != nil {
		return nil, err
	} else if b == '}' {
		// Empty object
		_, err := p.scanner.next()
		if err != nil {
			return nil, err
		}
		buf.writeByte('}')
		return buf.bytes(), nil
	}

	// Parse object content
	first := true
	for {
		if !first {
			// Expect comma or closing brace
			if err := p.scanner.skipWhitespace(); err != nil {
				return nil, err
			}

			b, err := p.scanner.next()
			if err != nil {
				return nil, err
			}

			if b == '}' {
				buf.writeByte('}')
				return buf.bytes(), nil
			} else if b == ',' {
				buf.writeByte(',')
			} else {
				return nil, newSyntaxError(p.scanner.position(), "expected ',' or '}'")
			}
		}
		first = false

		// Parse key-value pair
		if err := p.parseKeyValuePair(buf); err != nil {
			return nil, err
		}
	}
}

// parseArray parses a JSON array
func (p *parser) parseArray(buf *buffer) ([]byte, error) {
	p.depth++
	defer func() { p.depth-- }()

	if err := p.checkDepth(); err != nil {
		return nil, err
	}

	buf.writeByte('[')

	// Consume the opening bracket
	b, err := p.scanner.next()
	if err != nil {
		return nil, err
	}
	if b != '[' {
		return nil, newSyntaxError(p.scanner.position(), "expected '['")
	}

	// Skip whitespace
	if err := p.scanner.skipWhitespace(); err != nil {
		return nil, err
	}

	// Check for empty array
	if b, err := p.scanner.peek(); err != nil {
		return nil, err
	} else if b == ']' {
		// Empty array
		_, err := p.scanner.next()
		if err != nil {
			return nil, err
		}
		buf.writeByte(']')
		return buf.bytes(), nil
	}

	// Parse array content
	first := true
	for {
		if !first {
			// Expect comma or closing bracket
			if err := p.scanner.skipWhitespace(); err != nil {
				return nil, err
			}

			b, err := p.scanner.next()
			if err != nil {
				return nil, err
			}

			if b == ']' {
				buf.writeByte(']')
				return buf.bytes(), nil
			} else if b == ',' {
				buf.writeByte(',')
			} else {
				return nil, newSyntaxError(p.scanner.position(), "expected ',' or ']'")
			}
		}
		first = false

		// Parse array element
		if err := p.parseElement(buf); err != nil {
			return nil, err
		}
	}
}

// parseKeyValuePair parses a key-value pair in an object
func (p *parser) parseKeyValuePair(buf *buffer) error {
	// Skip whitespace before key
	if err := p.scanner.skipWhitespace(); err != nil {
		return err
	}

	// Parse key (must be a string)
	if err := p.parseString(buf); err != nil {
		return err
	}

	// Skip whitespace before colon
	if err := p.scanner.skipWhitespace(); err != nil {
		return err
	}

	// Expect colon
	b, err := p.scanner.next()
	if err != nil {
		return err
	}
	if b != ':' {
		return newSyntaxError(p.scanner.position(), "expected ':'")
	}
	buf.writeByte(':')

	// Skip whitespace after colon
	if err := p.scanner.skipWhitespace(); err != nil {
		return err
	}

	// Parse value
	return p.parseElement(buf)
}

// parseElement parses any JSON element
func (p *parser) parseElement(buf *buffer) error {
	if err := p.scanner.skipWhitespace(); err != nil {
		return err
	}

	b, err := p.scanner.peek()
	if err != nil {
		return err
	}

	switch b {
	case '{':
		// Nested object
		nestedBuf := getBuffer()
		defer putBuffer(nestedBuf)
		objBytes, err := p.parseObject(nestedBuf)
		if err != nil {
			return err
		}
		buf.write(objBytes)
		return nil
	case '[':
		// Nested array
		nestedBuf := getBuffer()
		defer putBuffer(nestedBuf)
		arrBytes, err := p.parseArray(nestedBuf)
		if err != nil {
			return err
		}
		buf.write(arrBytes)
		return nil
	case '"':
		// String
		return p.parseString(buf)
	case 't', 'f':
		// Boolean
		return p.parseBoolean(buf)
	case 'n':
		// Null
		return p.parseNull(buf)
	default:
		if (b >= '0' && b <= '9') || b == '-' {
			// Number
			return p.parseNumber(buf)
		}
		return newSyntaxError(p.scanner.position(), "unexpected character")
	}
}

// parseString parses a JSON string
func (p *parser) parseString(buf *buffer) error {
	buf.writeByte('"')

	// Consume opening quote
	b, err := p.scanner.next()
	if err != nil {
		return err
	}
	if b != '"' {
		return newSyntaxError(p.scanner.position(), "expected '\"'")
	}

	for {
		b, err := p.scanner.next()
		if err != nil {
			return err
		}

		if b == '"' {
			// Check if this quote is escaped by looking backwards
			// For robust parsing, we treat unescaped quotes as string terminators
			// but escaped quotes as part of the string content

			// Simple heuristic: if we haven't seen a backslash immediately before this,
			// treat it as string terminator. For more sophisticated parsing,
			// we'd need to track escape state properly.
			buf.writeByte('"')
			return nil
		}

		if b == '\\' {
			// Escape sequence - decode according to RFC 8259
			nextByte, err := p.scanner.next()
			if err != nil {
				return err
			}

			switch nextByte {
			case '"':
				buf.writeByte('\\')
				buf.writeByte('"')
			case '\\':
				buf.writeByte('\\')
				buf.writeByte('\\')
			case '/':
				buf.writeByte('/')
			case 'b':
				buf.writeByte('\\')
				buf.writeByte('b')
			case 'f':
				buf.writeByte('\\')
				buf.writeByte('f')
			case 'n':
				buf.writeByte('\\')
				buf.writeByte('n')
			case 'r':
				buf.writeByte('\\')
				buf.writeByte('r')
			case 't':
				buf.writeByte('\\')
				buf.writeByte('t')
			case 'u':
				// Unicode escape sequence - preserve as-is for now
				buf.writeByte('\\')
				buf.writeByte('u')
				for i := 0; i < 4; i++ {
					hexByte, err := p.scanner.next()
					if err != nil {
						return err
					}
					if !isHexDigit(hexByte) {
						return newEscapeError(p.scanner.position(), "invalid hex digit in unicode escape")
					}
					buf.writeByte(hexByte)
				}
			default:
				return newEscapeError(p.scanner.position(), "invalid escape sequence")
			}
		} else {
			// Regular character
			if b >= 0x80 {
				// Multi-byte UTF-8 character - need to read the complete sequence
				sequence := []byte{b}

				// Determine sequence length based on first byte
				var seqLen int
				if b&0xE0 == 0xC0 {
					seqLen = 2
				} else if b&0xF0 == 0xE0 {
					seqLen = 3
				} else if b&0xF8 == 0xF0 {
					seqLen = 4
				} else {
					return newUnicodeError(p.scanner.position(), "invalid UTF-8 start byte")
				}

				// Read remaining bytes of the sequence
				for i := 1; i < seqLen; i++ {
					nextByte, err := p.scanner.next()
					if err != nil {
						return err
					}
					if nextByte&0xC0 != 0x80 {
						return newUnicodeError(p.scanner.position(), "invalid UTF-8 continuation byte")
					}
					sequence = append(sequence, nextByte)
				}

				// Write the complete UTF-8 sequence
				buf.write(sequence)
			} else {
				// ASCII character - handle control characters
				if b < 0x20 {
					// Control character - convert to escape sequence
					switch b {
					case '\n':
						buf.writeByte('\\')
						buf.writeByte('n')
					case '\t':
						buf.writeByte('\\')
						buf.writeByte('t')
					case '\r':
						buf.writeByte('\\')
						buf.writeByte('r')
					case '\b':
						buf.writeByte('\\')
						buf.writeByte('b')
					case '\f':
						buf.writeByte('\\')
						buf.writeByte('f')
					default:
						// Other control characters as \uXXXX
						buf.writeByte('\\')
						buf.writeByte('u')
						buf.writeByte('0')
						buf.writeByte('0')
						if b < 16 {
							buf.writeByte('0')
						} else {
							buf.writeByte('1')
						}
						hexDigit := b % 16
						if hexDigit < 10 {
							buf.writeByte('0' + hexDigit)
						} else {
							buf.writeByte('A' + hexDigit - 10)
						}
					}
				} else {
					buf.writeByte(b)
				}
			}
		}
	}
}

// parseBoolean parses true or false
func (p *parser) parseBoolean(buf *buffer) error {
	b, err := p.scanner.peek()
	if err != nil {
		return err
	}

	if b == 't' {
		// Parse "true"
		expected := "true"
		for _, char := range expected {
			b, err := p.scanner.next()
			if err != nil {
				return err
			}
			if b != byte(char) {
				return newSyntaxError(p.scanner.position(), "invalid boolean value")
			}
			buf.writeByte(b)
		}
	} else if b == 'f' {
		// Parse "false"
		expected := "false"
		for _, char := range expected {
			b, err := p.scanner.next()
			if err != nil {
				return err
			}
			if b != byte(char) {
				return newSyntaxError(p.scanner.position(), "invalid boolean value")
			}
			buf.writeByte(b)
		}
	} else {
		return newSyntaxError(p.scanner.position(), "expected boolean value")
	}

	return nil
}

// parseNull parses null
func (p *parser) parseNull(buf *buffer) error {
	expected := "null"
	for _, char := range expected {
		b, err := p.scanner.next()
		if err != nil {
			return err
		}
		if b != byte(char) {
			return newSyntaxError(p.scanner.position(), "invalid null value")
		}
		buf.writeByte(b)
	}
	return nil
}

// parseNumber parses a JSON number
func (p *parser) parseNumber(buf *buffer) error {
	for {
		b, err := p.scanner.peek()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Check if character is part of a number
		if (b >= '0' && b <= '9') || b == '-' || b == '+' || b == '.' || b == 'e' || b == 'E' {
			b, err := p.scanner.next()
			if err != nil {
				return err
			}
			buf.writeByte(b)
		} else {
			// End of number
			break
		}
	}
	return nil
}

// checkDepth validates nesting depth against limits
func (p *parser) checkDepth() error {
	if p.depth >= p.options.maxDepth {
		return newSyntaxError(p.scanner.position(), "maximum nesting depth exceeded")
	}
	return nil
}
