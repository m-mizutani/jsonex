package jsonex

import (
	"bytes"
	"testing"
)

func TestProcessEscape(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
		hasError bool
	}{
		// No escapes
		{[]byte("hello"), []byte("hello"), false},
		{[]byte(""), []byte(""), false},

		// Basic escapes
		{[]byte(`hello\"world`), []byte(`hello"world`), false},
		{[]byte(`line1\\line2`), []byte(`line1\line2`), false},
		{[]byte(`path\/file`), []byte(`path/file`), false},
		{[]byte(`bell\b`), []byte("bell\b"), false},
		{[]byte(`form\f`), []byte("form\f"), false},
		{[]byte(`new\nline`), []byte("new\nline"), false},
		{[]byte(`carriage\rreturn`), []byte("carriage\rreturn"), false},
		{[]byte(`tab\there`), []byte("tab\there"), false},

		// Unicode escapes
		{[]byte(`\u0041`), []byte("A"), false},
		{[]byte(`\u3042`), []byte("あ"), false},

		// Surrogate pairs (😀 emoji)
		{[]byte(`\uD83D\uDE00`), []byte("😀"), false},

		// Error cases
		{[]byte(`\`), nil, true},           // Incomplete escape
		{[]byte(`\x`), nil, true},          // Invalid escape
		{[]byte(`\u123`), nil, true},       // Incomplete unicode
		{[]byte(`\uXXXX`), nil, true},      // Invalid hex
		{[]byte(`\uD83D\u1234`), nil, true}, // Invalid surrogate pair
		{[]byte(`\uDC00`), nil, true},      // Unexpected low surrogate
	}

	for _, test := range tests {
		result, err := processEscape(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("processEscape(%s) error = %v, expected error = %v", test.input, err, test.hasError)
			continue
		}
		if !test.hasError && !bytes.Equal(result, test.expected) {
			t.Errorf("processEscape(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestDecodeUnicodeEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected rune
		hasError bool
	}{
		{"0041", 'A', false},
		{"3042", 'あ', false},
		{"D83D", 0xD83D, false},
		{"de00", 0xDE00, false},
		{"FFFF", 0xFFFF, false},

		// Error cases
		{"123", 0, true},    // Too short
		{"12345", 0, true},  // Too long
		{"XXXX", 0, true},   // Invalid hex
		{"123G", 0, true},   // Invalid hex character
	}

	for _, test := range tests {
		result, err := decodeUnicodeEscape(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("decodeUnicodeEscape(%s) error = %v, expected error = %v", test.input, err, test.hasError)
			continue
		}
		if !test.hasError && result != test.expected {
			t.Errorf("decodeUnicodeEscape(%s) = 0x%X, expected 0x%X", test.input, result, test.expected)
		}
	}
}

func TestHasEscapeSequences(t *testing.T) {
	tests := []struct {
		input    []byte
		expected bool
	}{
		{[]byte("hello"), false},
		{[]byte(`hello\"world`), true},
		{[]byte(`no escapes here`), false},
		{[]byte(`\n\t\r`), true},
		{[]byte(""), false},
	}

	for _, test := range tests {
		result := hasEscapeSequences(test.input)
		if result != test.expected {
			t.Errorf("hasEscapeSequences(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsHexDigit(t *testing.T) {
	tests := []struct {
		input    byte
		expected bool
	}{
		{'0', true},
		{'9', true},
		{'A', true},
		{'F', true},
		{'a', true},
		{'f', true},
		{'G', false},
		{'z', false},
		{' ', false},
	}

	for _, test := range tests {
		result := isHexDigit(test.input)
		if result != test.expected {
			t.Errorf("isHexDigit(%c) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestEncodeEscape(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
	}{
		{[]byte("hello"), []byte("hello")},
		{[]byte(`"quoted"`), []byte(`\"quoted\"`)},
		{[]byte("back\\slash"), []byte("back\\\\slash")},
		{[]byte("new\nline"), []byte("new\\nline")},
		{[]byte("tab\there"), []byte("tab\\there")},
		{[]byte("\x01"), []byte("\\u1")}, // Control character
	}

	for _, test := range tests {
		result := encodeEscape(test.input)
		if !bytes.Equal(result, test.expected) {
			t.Errorf("encodeEscape(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestValidateEscapeSequence(t *testing.T) {
	tests := []struct {
		input    []byte
		pos      int
		hasError bool
	}{
		{[]byte(`\"`), 0, false},
		{[]byte(`\\`), 0, false},
		{[]byte(`\n`), 0, false},
		{[]byte(`\u1234`), 0, false},
		{[]byte(`hello\nworld`), 5, false},

		// Error cases
		{[]byte(`hello`), 0, true},       // Not an escape
		{[]byte(`\`), 0, true},           // Incomplete
		{[]byte(`\x`), 0, true},          // Invalid escape
		{[]byte(`\u123`), 0, true},       // Incomplete unicode
		{[]byte(`\uXXXX`), 0, true},      // Invalid hex
	}

	for _, test := range tests {
		err := validateEscapeSequence(test.input, test.pos)
		if (err != nil) != test.hasError {
			t.Errorf("validateEscapeSequence(%s, %d) error = %v, expected error = %v", 
				test.input, test.pos, err, test.hasError)
		}
	}
}

func TestCountEscapeSequences(t *testing.T) {
	tests := []struct {
		input    []byte
		expected int
	}{
		{[]byte("hello"), 0},
		{[]byte(`hello\"world`), 1},
		{[]byte(`\n\t\r`), 3},
		{[]byte(`\u1234\u5678`), 2},
		{[]byte(`mix\"ed\n\u1234`), 3},
		{[]byte(""), 0},
	}

	for _, test := range tests {
		result := countEscapeSequences(test.input)
		if result != test.expected {
			t.Errorf("countEscapeSequences(%s) = %d, expected %d", test.input, result, test.expected)
		}
	}
}