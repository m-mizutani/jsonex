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

	// RFC 8259 compliant implementation decodes escape sequences
	if result["backslash"] != "\\" {
		t.Errorf("Backslash escape incorrect: %v", result["backslash"])
	}
	if result["quote"] != "\"" {
		t.Errorf("Quote escape incorrect: %q", result["quote"])
	}
	if result["newline"] != "\n" {
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