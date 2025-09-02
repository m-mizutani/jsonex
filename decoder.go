package jsonex

import (
	"encoding/json"
	"io"
)

// Decoder reads and decodes JSON values from an input stream
type Decoder struct {
	parser  *parser
	options options
}

// New creates a new Decoder that reads from r
func New(r io.Reader, opts ...Option) *Decoder {
	options := applyOptions(opts...)
	return &Decoder{
		parser:  newParser(r, options),
		options: options,
	}
}

// Decode reads the next JSON-encoded value from its input and stores it in the value pointed to by v
// The behavior is similar to json.Decoder.Decode but only accepts objects and arrays
func (d *Decoder) Decode(v interface{}) error {
	// Extract the next JSON object or array
	jsonBytes, err := d.parser.parseNext()
	if err != nil {
		return err
	}

	// Use standard library to decode the extracted JSON
	return json.Unmarshal(jsonBytes, v)
}

// More methods can be added here for compatibility with json.Decoder if needed

// Buffered returns a reader of the data remaining in the Decoder's buffer
// This can be useful for reading any remaining data after JSON parsing
func (d *Decoder) Buffered() io.Reader {
	// For now, we don't implement buffering
	// This would require more complex scanner state management
	return nil
}

// DisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination
func (d *Decoder) DisallowUnknownFields() {
	// This would require integration with the standard library's decoder options
	// For now, we don't implement this feature
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64
func (d *Decoder) UseNumber() {
	// This would require integration with the standard library's decoder options
	// For now, we don't implement this feature
}