package jsonex

import (
	"bytes"
	"encoding/json"
)

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v
// Unlike the standard json.Unmarshal, this function extracts the longest valid JSON
// object or array from the input data, ignoring any preceding or trailing invalid content
func Unmarshal(data []byte, v interface{}, opts ...Option) error {
	if len(data) == 0 {
		return newInvalidJSONError(position{}, "empty input data")
	}

	options := applyOptions(opts...)

	// Fast path: try standard library first if data looks clean and no special options
	if options.maxDepth == 1000 && options.bufferSize == 4096 { // Default options only
		trimmed := bytes.TrimSpace(data)
		if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
			// Check if the trimmed data equals the original data (no garbage)
			if bytes.Equal(trimmed, data) {
				if err := json.Unmarshal(trimmed, v); err == nil {
					return nil
				}
			}
		}
	}

	// Robust path: find and extract the longest valid JSON
	jsonBytes, err := parseLongest(data, options)
	if err != nil {
		return err
	}

	// Use standard library to decode the extracted JSON
	// The standard library already handles all RFC 8259 compliant escape sequences
	return json.Unmarshal(jsonBytes, v)
}
