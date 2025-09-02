package jsonex

import (
	"strings"
	"testing"
)

func TestScanner_EmptyInput(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "Empty slice",
			data:        []byte{},
			expectError: true,
		},
		{
			name:        "Only whitespace",
			data:        []byte("   \t\n\r   "),
			expectError: true,
		},
		{
			name:        "Only garbage",
			data:        []byte("random text with no JSON"),
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestScanner_MinimalJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		desc string
	}{
		{
			name: "Empty object",
			data: []byte(`{}`),
			desc: "Minimal object",
		},
		{
			name: "Empty array",
			data: []byte(`[]`),
			desc: "Minimal array",
		},
		{
			name: "Object with garbage",
			data: []byte(`prefix {} suffix`),
			desc: "Empty object with surrounding garbage",
		},
		{
			name: "Array with garbage",
			data: []byte(`noise [] end`),
			desc: "Empty array with surrounding garbage",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("Minimal JSON test failed (%s): %v", test.desc, err)
			}
		})
	}
}

func TestScanner_MultipleJSONObjects(t *testing.T) {
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

func TestScanner_DecoderStreaming(t *testing.T) {
	// Test decoder with mixed content and multiple JSON objects
	input := `
		random text {"object": 1} more text
		[1, 2, 3] additional content
		{"final": true}
	`

	decoder := New(strings.NewReader(input))

	// Should find first object
	var obj1 map[string]interface{}
	err := decoder.Decode(&obj1)
	if err != nil {
		t.Fatalf("First decode failed: %v", err)
	}

	// Should find array
	var arr []interface{}
	err = decoder.Decode(&arr)
	if err != nil {
		t.Fatalf("Array decode failed: %v", err)
	}

	// Should find final object
	var obj2 map[string]interface{}
	err = decoder.Decode(&obj2)
	if err != nil {
		t.Fatalf("Final decode failed: %v", err)
	}

	if obj1["object"] != float64(1) {
		t.Errorf("First object incorrect: %v", obj1)
	}
	if len(arr) != 3 {
		t.Errorf("Array length incorrect: %d", len(arr))
	}
	if obj2["final"] != true {
		t.Errorf("Final object incorrect: %v", obj2)
	}
}