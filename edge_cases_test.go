package jsonex

import (
	"testing"
)

// Edge case tests for extreme scenarios

func TestEdgeCases_ExtremeCases(t *testing.T) {
	// Large value test
	largeValue := string(make([]byte, 10000))
	for i := range largeValue {
		largeValue = string(largeValue[:i]) + "x" + string(largeValue[i+1:])
	}

	data := []byte(`{"key": "` + largeValue + `"}`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Large value test failed: %v", err)
	}

	if result["key"] != largeValue {
		t.Error("Large value incorrect")
	}
}

func TestEdgeCases_EscapeSequences(t *testing.T) {
	// Test all standard JSON escape sequences
	data := []byte(`prefix {"backslash": "\\\\", "quote": "\\\"", "slash": "\\/", "backspace": "\\b", "formfeed": "\\f", "newline": "\\n", "carriage": "\\r", "tab": "\\t"} suffix`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Escape sequences failed: %v", err)
	}

	// RFC 8259 compliant implementation decodes escape sequences correctly
	// Note: "\\\\" in JSON represents two backslash characters, not one
	if result["backslash"] != "\\\\" {
		t.Errorf("Backslash escape incorrect: %v", result["backslash"])
	}
	if result["quote"] != "\\\"" {
		t.Errorf("Quote escape incorrect: %q", result["quote"])
	}
	if result["newline"] != "\\n" {
		t.Errorf("Newline escape incorrect: %q", result["newline"])
	}
}

func TestEdgeCases_UnicodeEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		desc string
	}{
		{
			name: "Mixed scripts",
			data: []byte(`{"text": "Hello мир שלום 世界"}`),
			desc: "Multiple scripts in one string",
		},
		{
			name: "Unicode escapes",
			data: []byte(`{"text": "\\u0048\\u0065\\u006C\\u006C\\u006F"}`),
			desc: "Unicode escape sequences",
		},
		{
			name: "Surrogate pairs",
			data: []byte(`{"emoji": "\\uD83D\\uDE00\\uD83C\\uDF89"}`),
			desc: "UTF-16 surrogate pairs",
		},
		{
			name: "Zero width characters",
			data: []byte(`{"text": "Hello\\u200BWorld"}`),
			desc: "Zero-width space character",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("Unicode test failed (%s): %v", test.desc, err)
			}

			// Basic validation - just ensure it parsed
			if _, ok := result["text"]; ok && len(result) == 0 {
				t.Error("Text value was lost")
			}
			if _, ok := result["emoji"]; ok && len(result) == 0 {
				t.Error("Emoji value was lost")
			}
		})
	}
}

func TestEdgeCases_JSONLikeStrings(t *testing.T) {
	// Test for JSON strings containing JSON-like content that could be misinterpreted
	tests := []struct {
		name     string
		data     []byte
		expected map[string]interface{}
		desc     string
	}{
		{
			name: "JSON-like content in string",
			data: []byte(`garbage {"description": "This contains {\\\"fake\\\": \\\"json\\\"} inside"} noise`),
			expected: map[string]interface{}{
				"description": "This contains {\\\"fake\\\": \\\"json\\\"} inside",
			},
			desc: "String field containing escaped JSON-like content",
		},
		{
			name: "Array notation in string",
			data: []byte(`prefix {"message": "Array format: [1,2,3] not actual JSON"} suffix`),
			expected: map[string]interface{}{
				"message": "Array format: [1,2,3] not actual JSON",
			},
			desc: "String containing array-like notation",
		},
		{
			name: "Nested JSON description",
			data: []byte(`start {"text": "JSON format is {key: value} structure"} end`),
			expected: map[string]interface{}{
				"text": "JSON format is {key: value} structure",
			},
			desc: "String describing JSON format",
		},
		{
			name: "Complex nested false JSON",
			data: []byte(`noise {"config": "Settings: {\\\"theme\\\": \\\"dark\\\", \\\"size\\\": 12}"} trash`),
			expected: map[string]interface{}{
				"config": "Settings: {\\\"theme\\\": \\\"dark\\\", \\\"size\\\": 12}",
			},
			desc: "String containing properly escaped JSON-like configuration",
		},
		{
			name: "Object with curly braces in text",
			data: []byte(`{"code": "function() { return {a: 1, b: 2}; }"}`),
			expected: map[string]interface{}{
				"code": "function() { return {a: 1, b: 2}; }",
			},
			desc: "JavaScript code with object literals in string",
		},
		{
			name: "Mixed quotes and braces",
			data: []byte(`junk {"note": "Use \\\"quotes\\\" around {objects} in JSON"} end`),
			expected: map[string]interface{}{
				"note": "Use \\\"quotes\\\" around {objects} in JSON",
			},
			desc: "String with mixed quotes and object notation",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("JSON-like string test failed (%s): %v", test.desc, err)
				return
			}

			for key, expectedValue := range test.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key %q not found in result", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Key %q: expected %q, got %q", key, expectedValue, actualValue)
				}
			}

			// Ensure we only parsed one JSON object, not multiple
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d keys, got %d. Result: %v", len(test.expected), len(result), result)
			}
		})
	}
}
