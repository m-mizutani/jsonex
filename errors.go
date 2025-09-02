package jsonex

import "fmt"

// ErrorType represents the type of error that occurred during parsing
type ErrorType int

const (
	ErrSyntax ErrorType = iota
	ErrUnicode
	ErrEscape
	ErrEOF
	ErrInvalidJSON
)

// String returns the string representation of ErrorType
func (t ErrorType) String() string {
	switch t {
	case ErrSyntax:
		return "syntax error"
	case ErrUnicode:
		return "unicode error"
	case ErrEscape:
		return "escape error"
	case ErrEOF:
		return "unexpected end of file"
	case ErrInvalidJSON:
		return "invalid json"
	default:
		return "unknown error"
	}
}

// Position represents a position in the input stream
type Position struct {
	Offset int // byte offset
	Line   int // line number (1-based)
	Column int // column number (1-based)
}

// String returns the string representation of Position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d (offset %d)", p.Line, p.Column, p.Offset)
}

// Error represents an error that occurred during JSON parsing
type Error struct {
	Type     ErrorType
	Message  string
	Position Position
	Context  string
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s at %s: %s (context: %s)", e.Type, e.Position, e.Message, e.Context)
	}
	return fmt.Sprintf("%s at %s: %s", e.Type, e.Position, e.Message)
}

// position represents internal position tracking (unexported)
type position struct {
	offset int
	line   int
	column int
}

// toPublic converts internal position to public Position
func (p position) toPublic() Position {
	return Position{
		Offset: p.offset,
		Line:   p.line,
		Column: p.column,
	}
}

// newError creates a new Error
func newError(t ErrorType, pos position, message string, context ...string) *Error {
	err := &Error{
		Type:     t,
		Message:  message,
		Position: pos.toPublic(),
	}
	if len(context) > 0 {
		err.Context = context[0]
	}
	return err
}

// newSyntaxError creates a new syntax error
func newSyntaxError(pos position, message string, context ...string) *Error {
	return newError(ErrSyntax, pos, message, context...)
}

// newUnicodeError creates a new unicode error
func newUnicodeError(pos position, message string, context ...string) *Error {
	return newError(ErrUnicode, pos, message, context...)
}

// newEscapeError creates a new escape error
func newEscapeError(pos position, message string, context ...string) *Error {
	return newError(ErrEscape, pos, message, context...)
}

// newEOFError creates a new EOF error
func newEOFError(pos position, message string, context ...string) *Error {
	return newError(ErrEOF, pos, message, context...)
}

// newInvalidJSONError creates a new invalid JSON error
func newInvalidJSONError(pos position, message string, context ...string) *Error {
	return newError(ErrInvalidJSON, pos, message, context...)
}
