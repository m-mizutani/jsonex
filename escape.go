package jsonex

import (
	"strconv"
	"strings"
)

// processEscape processes escape sequences in JSON strings
func processEscape(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	// Quick path: if no backslashes, return as-is
	if !hasEscapeSequences(data) {
		return data, nil
	}

	result := make([]byte, 0, len(data))
	pos := 0

	for pos < len(data) {
		if data[pos] != '\\' {
			result = append(result, data[pos])
			pos++
			continue
		}

		// Handle escape sequence
		if pos+1 >= len(data) {
			return nil, newEscapeError(position{offset: pos}, "incomplete escape sequence")
		}

		switch data[pos+1] {
		case '"':
			result = append(result, '"')
			pos += 2
		case '\\':
			result = append(result, '\\')
			pos += 2
		case '/':
			result = append(result, '/')
			pos += 2
		case 'b':
			result = append(result, '\b')
			pos += 2
		case 'f':
			result = append(result, '\f')
			pos += 2
		case 'n':
			result = append(result, '\n')
			pos += 2
		case 'r':
			result = append(result, '\r')
			pos += 2
		case 't':
			result = append(result, '\t')
			pos += 2
		case 'u':
			// Unicode escape sequence
			if pos+5 >= len(data) {
				return nil, newEscapeError(position{offset: pos}, "incomplete unicode escape sequence")
			}
			
			hexStr := string(data[pos+2 : pos+6])
			r, err := decodeUnicodeEscape(hexStr)
			if err != nil {
				return nil, newEscapeError(position{offset: pos}, "invalid unicode escape sequence: "+hexStr)
			}

			// Check for surrogate pairs
			if isHighSurrogate(r) {
				if pos+11 >= len(data) || data[pos+6] != '\\' || data[pos+7] != 'u' {
					return nil, newEscapeError(position{offset: pos}, "incomplete surrogate pair")
				}
				
				lowHexStr := string(data[pos+8 : pos+12])
				lowR, err := decodeUnicodeEscape(lowHexStr)
				if err != nil {
					return nil, newEscapeError(position{offset: pos}, "invalid low surrogate: "+lowHexStr)
				}
				
				if !isLowSurrogate(lowR) {
					return nil, newEscapeError(position{offset: pos}, "invalid surrogate pair")
				}
				
				// Decode surrogate pair
				codePoint := decodeSurrogatePair(r, lowR)
				utf8Bytes := encodeUTF8Rune(codePoint)
				result = append(result, utf8Bytes...)
				pos += 12
			} else if isLowSurrogate(r) {
				return nil, newEscapeError(position{offset: pos}, "unexpected low surrogate")
			} else {
				// Regular Unicode escape
				utf8Bytes := encodeUTF8Rune(r)
				result = append(result, utf8Bytes...)
				pos += 6
			}
		default:
			return nil, newEscapeError(position{offset: pos}, "invalid escape character: \\"+string(data[pos+1]))
		}
	}

	return result, nil
}

// decodeUnicodeEscape decodes a 4-character hex string to a rune
func decodeUnicodeEscape(hex string) (rune, error) {
	if len(hex) != 4 {
		return 0, newEscapeError(position{}, "unicode escape must be 4 characters")
	}

	// Check for valid hex characters
	for _, c := range hex {
		if !isHexDigit(byte(c)) {
			return 0, newEscapeError(position{}, "invalid hex character in unicode escape: "+string(c))
		}
	}

	value, err := strconv.ParseUint(hex, 16, 16)
	if err != nil {
		return 0, newEscapeError(position{}, "invalid unicode escape: "+hex)
	}

	return rune(value), nil
}

// hasEscapeSequences checks if the byte slice contains any escape sequences
func hasEscapeSequences(data []byte) bool {
	for _, b := range data {
		if b == '\\' {
			return true
		}
	}
	return false
}

// isHexDigit checks if a byte is a valid hexadecimal digit
func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'A' && b <= 'F') ||
		(b >= 'a' && b <= 'f')
}

// encodeEscape encodes special characters as escape sequences
func encodeEscape(data []byte) []byte {
	result := make([]byte, 0, len(data)*2) // Worst case: every byte needs escaping

	for _, b := range data {
		switch b {
		case '"':
			result = append(result, '\\', '"')
		case '\\':
			result = append(result, '\\', '\\')
		case '\b':
			result = append(result, '\\', 'b')
		case '\f':
			result = append(result, '\\', 'f')
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		case '\t':
			result = append(result, '\\', 't')
		default:
			if b < 0x20 {
				// Control characters need unicode escape
				result = append(result, []byte("\\u"+strings.ToUpper(strconv.FormatUint(uint64(b), 16)))...)
			} else {
				result = append(result, b)
			}
		}
	}

	return result
}

// validateEscapeSequence validates an escape sequence starting at the given position
func validateEscapeSequence(data []byte, pos int) error {
	if pos >= len(data) || data[pos] != '\\' {
		return newEscapeError(position{offset: pos}, "not an escape sequence")
	}

	if pos+1 >= len(data) {
		return newEscapeError(position{offset: pos}, "incomplete escape sequence")
	}

	switch data[pos+1] {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		return nil // Valid simple escape
	case 'u':
		if pos+5 >= len(data) {
			return newEscapeError(position{offset: pos}, "incomplete unicode escape")
		}
		// Validate hex digits
		for i := pos + 2; i < pos+6; i++ {
			if !isHexDigit(data[i]) {
				return newEscapeError(position{offset: pos}, "invalid hex digit in unicode escape")
			}
		}
		return nil
	default:
		return newEscapeError(position{offset: pos}, "invalid escape character")
	}
}

// countEscapeSequences counts the number of escape sequences in the data
func countEscapeSequences(data []byte) int {
	count := 0
	pos := 0
	for pos < len(data) {
		if data[pos] == '\\' && pos+1 < len(data) {
			count++
			// Skip the escape sequence
			if data[pos+1] == 'u' {
				pos += 6 // \uXXXX
			} else {
				pos += 2 // \X
			}
		} else {
			pos++
		}
	}
	return count
}