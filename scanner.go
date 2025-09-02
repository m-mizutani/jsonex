package jsonex

import (
	"io"
	"unicode/utf8"
)

// scanner handles low-level byte stream processing (unexported)
type scanner struct {
	reader io.Reader
	buffer []byte
	pos    int
	size   int
	line   int
	column int
	offset int
	eof    bool
}

// newScanner creates a new scanner
func newScanner(reader io.Reader, bufferSize int) *scanner {
	return &scanner{
		reader: reader,
		buffer: make([]byte, bufferSize),
		pos:    0,
		size:   0,
		line:   1,
		column: 1,
		offset: 0,
		eof:    false,
	}
}

// fillBuffer reads more data from the reader
func (s *scanner) fillBuffer() error {
	if s.eof {
		return io.EOF
	}

	// Move remaining bytes to the beginning
	if s.pos > 0 && s.pos < s.size {
		copy(s.buffer, s.buffer[s.pos:s.size])
		s.size -= s.pos
		s.pos = 0
	} else if s.pos >= s.size {
		s.size = 0
		s.pos = 0
	}

	// Read new data
	n, err := s.reader.Read(s.buffer[s.size:])
	s.size += n

	if err == io.EOF {
		s.eof = true
		if s.size == 0 {
			return io.EOF
		}
		return nil
	}
	return err
}

// peek returns the current byte without advancing
func (s *scanner) peek() (byte, error) {
	if s.pos >= s.size {
		if err := s.fillBuffer(); err != nil {
			return 0, err
		}
	}
	if s.pos >= s.size {
		return 0, io.EOF
	}
	return s.buffer[s.pos], nil
}

// next returns the current byte and advances the position
func (s *scanner) next() (byte, error) {
	if s.pos >= s.size {
		if err := s.fillBuffer(); err != nil {
			return 0, err
		}
	}
	if s.pos >= s.size {
		return 0, io.EOF
	}

	b := s.buffer[s.pos]
	s.pos++
	s.offset++

	// Update line and column tracking
	if b == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}

	return b, nil
}

// position returns the current position
func (s *scanner) position() position {
	return position{
		offset: s.offset,
		line:   s.line,
		column: s.column,
	}
}

// skipWhitespace skips whitespace characters (space, tab, newline, carriage return)
func (s *scanner) skipWhitespace() error {
	for {
		b, err := s.peek()
		if err != nil {
			return err
		}
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			break
		}
		_, err = s.next()
		if err != nil {
			return err
		}
	}
	return nil
}

// findJSONStart searches for the start of a JSON object or array
func (s *scanner) findJSONStart() (byte, error) {
	for {
		err := s.skipWhitespace()
		if err != nil {
			return 0, err
		}

		b, err := s.peek()
		if err != nil {
			return 0, err
		}

		// Check for JSON start characters (only objects and arrays)
		if b == '{' || b == '[' {
			return b, nil
		}

		// Skip invalid characters and continue searching
		_, err = s.next()
		if err != nil {
			return 0, err
		}
	}
}

// validateUTF8Byte checks if the current byte is part of a valid UTF-8 sequence
func (s *scanner) validateUTF8Byte(b byte) error {
	if b < 0x80 {
		// ASCII character, always valid
		return nil
	}

	// Multi-byte UTF-8 character
	// Read the full sequence to validate
	var sequence []byte

	// Determine sequence length based on first byte
	var seqLen int
	if b&0xE0 == 0xC0 {
		seqLen = 2
	} else if b&0xF0 == 0xE0 {
		seqLen = 3
	} else if b&0xF8 == 0xF0 {
		seqLen = 4
	} else {
		return newUnicodeError(s.position(), "invalid UTF-8 start byte")
	}

	// Collect the full sequence
	sequence = append(sequence, b)
	for i := 1; i < seqLen; i++ {
		nextByte, err := s.peek()
		if err != nil {
			return newUnicodeError(s.position(), "incomplete UTF-8 sequence")
		}
		if nextByte&0xC0 != 0x80 {
			return newUnicodeError(s.position(), "invalid UTF-8 continuation byte")
		}
		sequence = append(sequence, nextByte)
		_, err = s.next()
		if err != nil {
			return err
		}
	}

	// Validate the complete sequence
	if !utf8.Valid(sequence) {
		return newUnicodeError(s.position(), "invalid UTF-8 sequence")
	}

	return nil
}

// readBytes reads exactly n bytes from the current position
func (s *scanner) readBytes(n int) ([]byte, error) {
	result := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		b, err := s.next()
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

// unread moves back one position (limited capability)
func (s *scanner) unread() {
	if s.pos > 0 {
		s.pos--
		s.offset--
		// Note: This doesn't properly handle line/column tracking
		// Should only be used in specific cases where we know the previous character
	}
}
