package jsonex

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzUnmarshal tests the Unmarshal function with various inputs
func FuzzUnmarshal(f *testing.F) {
	// Add seed corpus with known valid JSON patterns
	f.Add([]byte(`{"key": "value"}`))
	f.Add([]byte(`[1, 2, 3]`))
	f.Add([]byte(`{"nested": {"deep": {"object": true}}}`))
	f.Add([]byte(`"simple string"`))
	f.Add([]byte(`123.456`))
	f.Add([]byte(`true`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"unicode": "こんにちは"}`))
	f.Add([]byte(`{"escape": "line1\nline2\ttab"}`))
	f.Add([]byte(`{"array": [{"nested": true}, 42, "string"]}`))
	
	// Add seed corpus with JSON embedded in noise (testing robust parsing)
	f.Add([]byte(`garbage {"valid": "json"} more garbage`))
	f.Add([]byte(`prefix noise [1,2,3] suffix noise`))
	f.Add([]byte(`{"mixed": true} and {"more": "data"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test with various target types
		var (
			mapTarget    map[string]interface{}
			sliceTarget  []interface{}
			anyTarget    interface{}
			stringTarget string
		)
		testTargets := []interface{}{
			&mapTarget,
			&sliceTarget,
			&anyTarget,
			&stringTarget,
		}

		for _, target := range testTargets {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Unmarshal panicked with input %q: %v", data, r)
					}
				}()

				err := Unmarshal(data, target)
				if err != nil {
					// Error is acceptable, just ensure it's a proper error type
					if !isAcceptableError(err) {
						t.Errorf("Unexpected error type for input %q: %T: %v", data, err, err)
					}
				}
			}()
		}
	})
}

// FuzzDecoder tests the Decoder.Decode function with streaming inputs
func FuzzDecoder(f *testing.F) {
	// Add seed corpus for streaming scenarios
	f.Add([]byte(`{"first": 1} {"second": 2}`))
	f.Add([]byte(`[1,2,3] [4,5,6]`))
	f.Add([]byte(`"string1" "string2" "string3"`))
	f.Add([]byte(`garbage {"valid": 1} noise {"another": 2} end`))
	f.Add([]byte(`true false null`))
	f.Add([]byte(`{"incomplete": "data`)) // Incomplete JSON
	f.Add([]byte(`{"deep": {"nested": {"object": {"with": {"many": {"levels": true}}}}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Decoder panicked with input %q: %v", data, r)
			}
		}()

		reader := strings.NewReader(string(data))
		decoder := New(reader)

		// Try to decode multiple objects from the stream
		for i := 0; i < 10; i++ { // Limit iterations to prevent infinite loops
			var result interface{}
			err := decoder.Decode(&result)
			if err != nil {
				// EOF or other errors are acceptable for streaming
				break
			}
		}
	})
}

// FuzzUnicodeHandling tests Unicode-specific processing
func FuzzUnicodeHandling(f *testing.F) {
	// Add seed corpus with various Unicode patterns
	f.Add([]byte(`{"unicode": "Hello 世界"}`))
	f.Add([]byte(`{"emoji": "🚀🌟⭐"}`))
	f.Add([]byte(`{"escape": "\u0048\u0065\u006c\u006c\u006f"}`)) // "Hello" in Unicode escapes
	f.Add([]byte(`{"mixed": "ASCII混合unicode🎉"}`))
	f.Add([]byte(`{"surrogate": "\uD83D\uDE00"}`)) // 😀 emoji using surrogate pair
	f.Add([]byte(`{"control": "\u0000\u0001\u0002"}`)) // Control characters
	f.Add([]byte(`{"newlines": "line1\u000Aline2\u000D\u000Aline3"}`))
	f.Add([]byte(`{"tab": "col1\u0009col2\u0009col3"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unicode handling panicked with input %q: %v", data, r)
			}
		}()

		// Ensure input is valid UTF-8 or test our handling of invalid sequences
		if !utf8.Valid(data) {
			// For invalid UTF-8, we still shouldn't panic
			var result interface{}
			err := Unmarshal(data, &result)
			if err != nil && !isAcceptableError(err) {
				t.Errorf("Unexpected error for invalid UTF-8 %q: %v", data, err)
			}
			return
		}

		var result map[string]interface{}
		err := Unmarshal(data, &result)
		if err != nil && !isAcceptableError(err) {
			t.Errorf("Unexpected error for Unicode input %q: %v", data, err)
		}
	})
}

// FuzzDeepNesting tests deeply nested JSON structures
func FuzzDeepNesting(f *testing.F) {
	// Add seed corpus with various nesting patterns
	f.Add(generateNestedObject(5))
	f.Add(generateNestedArray(5))
	f.Add(generateMixedNesting(3))
	f.Add(generateNestedObject(10))
	f.Add(generateNestedArray(10))
	f.Add(generateMixedNesting(7))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Deep nesting panicked with input %q: %v", data, r)
			}
		}()

		// Test with depth limits
		depthLimits := []int{5, 10, 50, 100}
		
		for _, limit := range depthLimits {
			var result interface{}
			err := Unmarshal(data, &result, WithMaxDepth(limit))
			if err != nil && !isAcceptableError(err) {
				t.Errorf("Unexpected error for depth limit %d with input %q: %v", limit, data, err)
			}
		}
	})
}

// Helper functions for fuzzing

// isAcceptableError checks if an error is expected/acceptable during fuzzing
func isAcceptableError(err error) bool {
	if err == nil {
		return true
	}

	// Check for our custom error types
	switch err.(type) {
	case *Error:
		return true
	case *json.SyntaxError, *json.UnmarshalTypeError:
		return true
	}

	// Check for common error patterns in error messages
	errMsg := err.Error()
	acceptablePatterns := []string{
		"invalid JSON",
		"syntax error",
		"unexpected",
		"EOF",
		"invalid character",
		"invalid Unicode",
		"depth limit",
		"buffer size",
		"malformed",
	}

	for _, pattern := range acceptablePatterns {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// generateNestedObject creates a deeply nested JSON object
func generateNestedObject(depth int) []byte {
	if depth <= 0 {
		return []byte(`{"value": "leaf"}`)
	}
	
	inner := generateNestedObject(depth - 1)
	return []byte(`{"level": ` + string(inner) + `}`)
}

// generateNestedArray creates a deeply nested JSON array
func generateNestedArray(depth int) []byte {
	if depth <= 0 {
		return []byte(`["leaf"]`)
	}
	
	inner := generateNestedArray(depth - 1)
	return []byte(`[` + string(inner) + `]`)
}

// generateMixedNesting creates mixed object/array nesting
func generateMixedNesting(depth int) []byte {
	if depth <= 0 {
		return []byte(`"leaf"`)
	}
	
	inner := generateMixedNesting(depth - 1)
	if depth%2 == 0 {
		return []byte(`{"nested": ` + string(inner) + `}`)
	} else {
		return []byte(`[` + string(inner) + `]`)
	}
}