package jsonex

import (
	"bytes"
	"encoding/json"
	"io"
)

// optimizedUnmarshal provides a faster path for clean JSON data without garbage
func optimizedUnmarshal(data []byte, v interface{}, opts options) error {
	// Quick check: if data starts with { or [, try standard library first
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		// Try standard library first (fast path)
		if err := json.Unmarshal(trimmed, v); err == nil {
			return nil
		}
	}
	
	// Fall back to robust parsing
	return robustUnmarshal(data, v, opts)
}

// robustUnmarshal is the original implementation for data with garbage
func robustUnmarshal(data []byte, v interface{}, opts options) error {
	jsonBytes, err := parseLongest(data, opts)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, v)
}

// optimizedDecoder provides faster decoding for clean streams
func (d *Decoder) optimizedDecode(v interface{}) error {
	// Try to peek ahead and see if we have clean JSON
	if reader, ok := d.parser.scanner.reader.(*bytes.Reader); ok {
		// For bytes.Reader, we can optimize
		return d.decodeFromBytesReader(reader, v)
	}
	
	// Fall back to robust parsing
	return d.robustDecode(v)
}

// decodeFromBytesReader optimizes decoding from bytes.Reader
func (d *Decoder) decodeFromBytesReader(reader *bytes.Reader, v interface{}) error {
	// Get remaining data
	remaining := make([]byte, reader.Len())
	reader.Read(remaining)
	reader.Seek(-int64(len(remaining)), io.SeekCurrent) // Reset position
	
	// Try fast path with standard library
	trimmed := bytes.TrimSpace(remaining)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		if err := decoder.Decode(v); err == nil {
			// Fast path succeeded, advance reader
			advance := len(remaining) - len(trimmed) + int(decoder.InputOffset())
			reader.Seek(int64(advance), io.SeekCurrent)
			return nil
		}
	}
	
	// Fall back to robust parsing
	return d.robustDecode(v)
}

// robustDecode is the original implementation
func (d *Decoder) robustDecode(v interface{}) error {
	jsonBytes, err := d.parser.parseNext()
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, v)
}

// Update the main functions to use optimized versions

// UnmarshalOptimized tries fast path first, then falls back to robust parsing
func UnmarshalOptimized(data []byte, v interface{}, opts ...Option) error {
	if len(data) == 0 {
		return newInvalidJSONError(position{}, "empty input data")
	}
	
	options := applyOptions(opts...)
	return optimizedUnmarshal(data, v, options)
}

// DecodeOptimized provides optimized decoding
func (d *Decoder) DecodeOptimized(v interface{}) error {
	return d.optimizedDecode(v)
}