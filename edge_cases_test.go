package jsonex

import (
	"strings"
	"testing"
)

// Additional edge case tests for maximum robustness

func TestEdgeCases_EmptyInput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty slice", []byte{}},
		{"Only whitespace", []byte("   \n\t\r   ")},
		{"Only garbage", []byte("garbage text with no JSON")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err == nil {
				t.Error("Expected error for empty/invalid input")
			}
		})
	}
}

func TestEdgeCases_MinimalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected interface{}
	}{
		{"Empty object", []byte("{}"), map[string]interface{}{}},
		{"Empty array", []byte("[]"), []interface{}{}},
		{"Object with garbage", []byte("trash {} more trash"), map[string]interface{}{}},
		{"Array with garbage", []byte("prefix [] suffix"), []interface{}{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestEdgeCases_MalformedJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Missing closing brace", []byte(`{"key": "value"`)},
		{"Missing closing bracket", []byte(`["item1", "item2"`)},
		{"Extra comma object", []byte(`{"key": "value",}`)},
		{"Extra comma array", []byte(`["item1", "item2",]`)},
		{"Missing quotes on key", []byte(`{key: "value"}`)},
		{"Single quotes", []byte(`{'key': 'value'}`)},
		{"Unescaped quotes in string", []byte(`{"key": "val"ue"}`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err == nil {
				t.Error("Expected error for malformed JSON")
			}
		})
	}
}

func TestEdgeCases_ExtremeCases(t *testing.T) {
	// Very large key
	largeKey := strings.Repeat("a", 10000)
	data := []byte(`{"` + largeKey + `": "value"}`)
	var result map[string]interface{}
	err := Unmarshal(data, &result)
	if err != nil {
		t.Errorf("Large key failed: %v", err)
	}
	if result[largeKey] != "value" {
		t.Error("Large key value incorrect")
	}

	// Very large string value
	largeValue := strings.Repeat("x", 10000)
	data2 := []byte(`{"key": "` + largeValue + `"}`)
	var result2 map[string]interface{}
	err = Unmarshal(data2, &result2)
	if err != nil {
		t.Errorf("Large value failed: %v", err)
	}
	if result2["key"] != largeValue {
		t.Error("Large value incorrect")
	}
}

func TestEdgeCases_MultipleJSONObjects(t *testing.T) {
	// Multiple complete JSON objects in sequence
	input := `noise {"first": 1} middle {"second": 2} {"third": 3} end`
	
	// Should find the longest valid JSON (our implementation finds longest, not first)
	var result map[string]interface{}
	err := Unmarshal([]byte(input), &result)
	if err != nil {
		t.Fatalf("Multiple JSON objects failed: %v", err)
	}
	
	// Should get one of the valid objects
	if result["first"] == float64(1) || result["second"] == float64(2) || result["third"] == float64(3) {
		// Any of these is valid
	} else {
		t.Errorf("Expected valid object, got: %v", result)
	}
}

func TestEdgeCases_NestedStructures(t *testing.T) {
	// Array containing objects containing arrays
	data := []byte(`garbage [
		{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]},
		{"users": [{"id": 3, "name": "Charlie"}]},
		{"empty": []}
	] suffix`)

	var result []interface{}
	err := Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Nested structures failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 top-level items, got %d", len(result))
	}

	// Verify nested structure
	firstItem := result[0].(map[string]interface{})
	users := firstItem["users"].([]interface{})
	if len(users) != 2 {
		t.Errorf("Expected 2 users in first item, got %d", len(users))
	}
}

func TestEdgeCases_UnicodeEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			"Mixed scripts",
			[]byte(`{"english": "hello", "japanese": "こんにちは", "arabic": "مرحبا", "emoji": "🌍🚀💻"}`),
		},
		{
			"Unicode escapes",
			[]byte(`{"escaped": "\\u0048\\u0065\\u006c\\u006c\\u006f"}`), // "Hello"
		},
		{
			"Surrogate pairs",
			[]byte(`{"emoji": "\\uD83D\\uDE00\\uD83C\\uDF89"}`), // 😀🎉
		},
		{
			"Zero width characters",
			[]byte(`{"text": "hello\\u200Bworld"}`), // Zero-width space
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("Unicode test failed: %v", err)
			}
		})
	}
}

func TestEdgeCases_NumberFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Integer", []byte(`{"num": 42}`)},
		{"Float", []byte(`{"num": 3.14}`)},
		{"Scientific notation", []byte(`{"num": 1.23e10}`)},
		{"Negative scientific", []byte(`{"num": -1.23E-10}`)},
		{"Zero", []byte(`{"num": 0}`)},
		{"Negative zero", []byte(`{"num": -0}`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("Number format test failed: %v", err)
			}
		})
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

func TestEdgeCases_DecoderStreaming(t *testing.T) {
	// Test decoder with mixed content and multiple JSON objects
	input := `
		random text
		{"first": {"nested": true}}
		more garbage
		[1, 2, {"array_nested": "value"}]
		final trash
		{"last": "object"}
	`

	decoder := New(strings.NewReader(input))

	// First object
	var obj1 map[string]interface{}
	err := decoder.Decode(&obj1)
	if err != nil {
		t.Fatalf("First decode failed: %v", err)
	}
	nested := obj1["first"].(map[string]interface{})
	if nested["nested"] != true {
		t.Error("First object incorrect")
	}

	// Second (array)
	var arr []interface{}
	err = decoder.Decode(&arr)
	if err != nil {
		t.Fatalf("Second decode failed: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("Array length incorrect: %d", len(arr))
	}

	// Third object
	var obj3 map[string]interface{}
	err = decoder.Decode(&obj3)
	if err != nil {
		t.Fatalf("Third decode failed: %v", err)
	}
	if obj3["last"] != "object" {
		t.Error("Last object incorrect")
	}
}

func TestEdgeCases_MaxDepthEnforcement(t *testing.T) {
	// Create deeply nested JSON that exceeds various depth limits
	baseJSON := `{"l1": {"l2": {"l3": {"l4": {"l5": {"l6": {"l7": {"l8": {"l9": {"l10": "deep"}}}}}}}}}`

	tests := []struct {
		name      string
		maxDepth  int
		shouldErr bool
	}{
		{"Depth 5 - should fail", 5, true},
		{"Depth 8 - should fail", 8, true},
		{"Depth 12 - should pass", 12, false},
		{"Depth 15 - should pass", 15, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal([]byte(baseJSON), &result, WithMaxDepth(test.maxDepth))

			if test.shouldErr && err == nil {
				t.Error("Expected depth error but got none")
			}
			if !test.shouldErr && err != nil {
				t.Errorf("Unexpected depth error: %v", err)
			}
		})
	}
}