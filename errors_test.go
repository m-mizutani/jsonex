package jsonex

import (
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrSyntax, "syntax error"},
		{ErrUnicode, "unicode error"},
		{ErrEscape, "escape error"},
		{ErrEOF, "unexpected end of file"},
		{ErrInvalidJSON, "invalid json"},
		{ErrorType(999), "unknown error"},
	}

	for _, test := range tests {
		result := test.errorType.String()
		if result != test.expected {
			t.Errorf("ErrorType(%d).String() = %s, expected %s", test.errorType, result, test.expected)
		}
	}
}

func TestPosition_String(t *testing.T) {
	pos := Position{Offset: 42, Line: 3, Column: 15}
	expected := "line 3, column 15 (offset 42)"
	result := pos.String()
	if result != expected {
		t.Errorf("Position.String() = %s, expected %s", result, expected)
	}
}

func TestError_Error(t *testing.T) {
	pos := Position{Offset: 10, Line: 2, Column: 5}

	// Test without context
	err := &Error{
		Type:     ErrSyntax,
		Message:  "unexpected character",
		Position: pos,
	}
	expected := "syntax error at line 2, column 5 (offset 10): unexpected character"
	result := err.Error()
	if result != expected {
		t.Errorf("Error.Error() = %s, expected %s", result, expected)
	}

	// Test with context
	err.Context = "parsing object"
	expected = "syntax error at line 2, column 5 (offset 10): unexpected character (context: parsing object)"
	result = err.Error()
	if result != expected {
		t.Errorf("Error.Error() with context = %s, expected %s", result, expected)
	}
}

func TestNewError(t *testing.T) {
	pos := position{offset: 5, line: 1, column: 6}

	err := newError(ErrSyntax, pos, "test message")
	if err.Type != ErrSyntax {
		t.Errorf("newError Type = %v, expected %v", err.Type, ErrSyntax)
	}
	if err.Message != "test message" {
		t.Errorf("newError Message = %s, expected %s", err.Message, "test message")
	}
	if err.Position.Offset != 5 || err.Position.Line != 1 || err.Position.Column != 6 {
		t.Errorf("newError Position = %v, expected {5 1 6}", err.Position)
	}
}

func TestNewErrorWithContext(t *testing.T) {
	pos := position{offset: 10, line: 2, column: 1}

	err := newError(ErrUnicode, pos, "invalid character", "test context")
	if err.Context != "test context" {
		t.Errorf("newError Context = %s, expected %s", err.Context, "test context")
	}
}
