package jsonex

import (
	"unicode/utf8"
)

// validateUTF8 checks if the given byte slice contains valid UTF-8
func validateUTF8(data []byte) error {
	if !utf8.Valid(data) {
		return newUnicodeError(position{}, "invalid UTF-8 sequence")
	}
	return nil
}

// decodeSurrogatePair converts a UTF-16 surrogate pair to a Unicode code point
func decodeSurrogatePair(high, low rune) rune {
	if !isHighSurrogate(high) || !isLowSurrogate(low) {
		return utf8.RuneError
	}
	return 0x10000 + (high-0xD800)<<10 + (low - 0xDC00)
}

// isHighSurrogate checks if the rune is a high surrogate
func isHighSurrogate(r rune) bool {
	return r >= 0xD800 && r <= 0xDBFF
}

// isLowSurrogate checks if the rune is a low surrogate
func isLowSurrogate(r rune) bool {
	return r >= 0xDC00 && r <= 0xDFFF
}

// isSurrogate checks if the rune is a surrogate (high or low)
func isSurrogate(r rune) bool {
	return r >= 0xD800 && r <= 0xDFFF
}

// isControlChar checks if the rune is a control character
func isControlChar(r rune) bool {
	return r >= 0x0000 && r <= 0x001F
}

// isValidUnicodeCodePoint checks if the rune is a valid Unicode code point
func isValidUnicodeCodePoint(r rune) bool {
	// Check for valid Unicode range
	if r < 0 || r > 0x10FFFF {
		return false
	}
	// Check for surrogate range (invalid in UTF-8)
	if isSurrogate(r) {
		return false
	}
	return true
}

// encodeUTF8Rune encodes a rune to UTF-8 bytes
func encodeUTF8Rune(r rune) []byte {
	if !isValidUnicodeCodePoint(r) {
		// Return replacement character for invalid code points
		return []byte{0xEF, 0xBF, 0xBD} // UTF-8 encoding of U+FFFD
	}

	var buf [4]byte
	n := utf8.EncodeRune(buf[:], r)
	return buf[:n]
}

// decodeUTF8Rune decodes the first rune from UTF-8 bytes
func decodeUTF8Rune(data []byte) (rune, int, error) {
	if len(data) == 0 {
		return 0, 0, newUnicodeError(position{}, "empty byte sequence")
	}

	r, size := utf8.DecodeRune(data)
	if r == utf8.RuneError && size == 1 {
		return 0, 0, newUnicodeError(position{}, "invalid UTF-8 sequence")
	}

	return r, size, nil
}

// normalizeUnicode normalizes Unicode text for consistent processing
func normalizeUnicode(data []byte) []byte {
	// For now, just validate and return as-is
	// Could implement Unicode normalization (NFC, NFD, etc.) if needed
	if validateUTF8(data) != nil {
		// Replace invalid sequences with replacement character
		return []byte("\uFFFD")
	}
	return data
}

// countRunes counts the number of runes in a UTF-8 byte slice
func countRunes(data []byte) int {
	return utf8.RuneCount(data)
}

// runeAt returns the rune at the specified index in a UTF-8 byte slice
func runeAt(data []byte, index int) (rune, error) {
	count := 0
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			return 0, newUnicodeError(position{}, "invalid UTF-8 sequence")
		}
		if count == index {
			return r, nil
		}
		count++
		data = data[size:]
	}
	return 0, newUnicodeError(position{}, "index out of range")
}

// hasValidUTF8Prefix checks if the byte slice starts with valid UTF-8
func hasValidUTF8Prefix(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	r, size := utf8.DecodeRune(data)
	return r != utf8.RuneError || size != 1
}
