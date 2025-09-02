package jsonex

import (
	"testing"
	"unicode/utf8"
)

func TestValidateUTF8(t *testing.T) {
	tests := []struct {
		input    []byte
		expected bool
	}{
		{[]byte("hello"), true},
		{[]byte("こんにちは"), true},
		{[]byte("🙂"), true},
		{[]byte{0xff, 0xfe}, false}, // Invalid UTF-8
		{[]byte{0xc0, 0x80}, false}, // Overlong encoding
	}

	for _, test := range tests {
		err := validateUTF8(test.input)
		if (err == nil) != test.expected {
			t.Errorf("validateUTF8(%v) error = %v, expected valid = %v", test.input, err, test.expected)
		}
	}
}

func TestDecodeSurrogatePair(t *testing.T) {
	// Test valid surrogate pair (😀 emoji)
	high := rune(0xD83D)
	low := rune(0xDE00)
	result := decodeSurrogatePair(high, low)
	expected := rune(0x1F600)

	if result != expected {
		t.Errorf("decodeSurrogatePair(0x%X, 0x%X) = 0x%X, expected 0x%X", high, low, result, expected)
	}

	// Test invalid surrogate pairs
	invalidTests := []struct {
		high, low rune
	}{
		{0x1234, 0xDC00}, // Invalid high surrogate
		{0xD800, 0x1234}, // Invalid low surrogate
		{0xDC00, 0xD800}, // Swapped order
	}

	for _, test := range invalidTests {
		result := decodeSurrogatePair(test.high, test.low)
		if result != utf8.RuneError {
			t.Errorf("decodeSurrogatePair(0x%X, 0x%X) = 0x%X, expected RuneError", test.high, test.low, result)
		}
	}
}

func TestIsHighSurrogate(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{0xD800, true},
		{0xDBFF, true},
		{0xDC00, false}, // Low surrogate
		{0x1234, false}, // Regular character
	}

	for _, test := range tests {
		result := isHighSurrogate(test.input)
		if result != test.expected {
			t.Errorf("isHighSurrogate(0x%X) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsLowSurrogate(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{0xDC00, true},
		{0xDFFF, true},
		{0xD800, false}, // High surrogate
		{0x1234, false}, // Regular character
	}

	for _, test := range tests {
		result := isLowSurrogate(test.input)
		if result != test.expected {
			t.Errorf("isLowSurrogate(0x%X) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsControlChar(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{0x00, true},
		{0x1F, true},
		{0x20, false}, // Space
		{0x7F, false}, // DEL (not in control range 0x00-0x1F)
		{'A', false},
	}

	for _, test := range tests {
		result := isControlChar(test.input)
		if result != test.expected {
			t.Errorf("isControlChar(0x%X) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsValidUnicodeCodePoint(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{'A', true},
		{0x1F600, true},   // Emoji
		{-1, false},       // Negative
		{0x110000, false}, // Too large
		{0xD800, false},   // Surrogate
		{0xDFFF, false},   // Surrogate
	}

	for _, test := range tests {
		result := isValidUnicodeCodePoint(test.input)
		if result != test.expected {
			t.Errorf("isValidUnicodeCodePoint(0x%X) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestEncodeUTF8Rune(t *testing.T) {
	tests := []struct {
		input    rune
		expected []byte
	}{
		{'A', []byte("A")},
		{0x1F600, []byte("😀")},
		{0xD800, []byte{0xEF, 0xBF, 0xBD}}, // Replacement character for invalid surrogate
	}

	for _, test := range tests {
		result := encodeUTF8Rune(test.input)
		if string(result) != string(test.expected) {
			t.Errorf("encodeUTF8Rune(0x%X) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestDecodeUTF8Rune(t *testing.T) {
	tests := []struct {
		input       []byte
		expectedR   rune
		expectedLen int
		expectError bool
	}{
		{[]byte("A"), 'A', 1, false},
		{[]byte("😀"), 0x1F600, 4, false},
		{[]byte{}, 0, 0, true},     // Empty
		{[]byte{0xFF}, 0, 0, true}, // Invalid UTF-8
	}

	for _, test := range tests {
		r, size, err := decodeUTF8Rune(test.input)
		if (err != nil) != test.expectError {
			t.Errorf("decodeUTF8Rune(%v) error = %v, expected error = %v", test.input, err, test.expectError)
			continue
		}
		if !test.expectError {
			if r != test.expectedR {
				t.Errorf("decodeUTF8Rune(%v) rune = 0x%X, expected 0x%X", test.input, r, test.expectedR)
			}
			if size != test.expectedLen {
				t.Errorf("decodeUTF8Rune(%v) size = %d, expected %d", test.input, size, test.expectedLen)
			}
		}
	}
}

func TestCountRunes(t *testing.T) {
	tests := []struct {
		input    []byte
		expected int
	}{
		{[]byte("hello"), 5},
		{[]byte("こんにちは"), 5},
		{[]byte("🙂🙃"), 2},
		{[]byte(""), 0},
	}

	for _, test := range tests {
		result := countRunes(test.input)
		if result != test.expected {
			t.Errorf("countRunes(%s) = %d, expected %d", test.input, result, test.expected)
		}
	}
}
