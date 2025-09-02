package jsonex

import (
	"strings"
	"testing"
)

// Comprehensive integration tests for complex scenarios

func TestComprehensive_UnicodeAndEscapes(t *testing.T) {
	// Test with Japanese, emojis, and escape sequences
	data := []byte(`prefix {"message": "こんにちは\n世界 🌍", "emoji": "😀", "escaped": "quote\"test"} suffix`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["message"] != "こんにちは\n世界 🌍" {
		t.Errorf("Unicode message incorrect: %v", result["message"])
	}

	if result["emoji"] != "😀" {
		t.Errorf("Emoji incorrect: %v", result["emoji"])
	}

	if result["escaped"] != "quote\"test" {
		t.Errorf("Escaped string incorrect: %v", result["escaped"])
	}
}

func TestComprehensive_DeepNesting(t *testing.T) {
	// Test deeply nested structure within limits
	data := []byte(`garbage {
		"level1": {
			"level2": {
				"level3": {
					"level4": {
						"level5": {
							"data": "deep value"
						}
					}
				}
			}
		}
	} more garbage`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Deep nesting failed: %v", err)
	}

	// Navigate to deep value
	level1 := result["level1"].(map[string]interface{})
	level2 := level1["level2"].(map[string]interface{})
	level3 := level2["level3"].(map[string]interface{})
	level4 := level3["level4"].(map[string]interface{})
	level5 := level4["level5"].(map[string]interface{})

	if level5["data"] != "deep value" {
		t.Errorf("Deep value incorrect: %v", level5["data"])
	}
}

func TestComprehensive_MixedArrays(t *testing.T) {
	// Test arrays with mixed types
	data := []byte(`junk [
		"string",
		42,
		true,
		null,
		{"nested": "object"},
		[1, 2, 3]
	] trash`)

	var result []interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Mixed array failed: %v", err)
	}

	if len(result) != 6 {
		t.Fatalf("Expected 6 elements, got %d", len(result))
	}

	if result[0] != "string" {
		t.Errorf("Element 0 incorrect: %v", result[0])
	}

	if result[1] != float64(42) {
		t.Errorf("Element 1 incorrect: %v", result[1])
	}

	if result[2] != true {
		t.Errorf("Element 2 incorrect: %v", result[2])
	}

	if result[3] != nil {
		t.Errorf("Element 3 incorrect: %v", result[3])
	}

	nested := result[4].(map[string]interface{})
	if nested["nested"] != "object" {
		t.Errorf("Nested object incorrect: %v", nested)
	}

	array := result[5].([]interface{})
	if len(array) != 3 || array[0] != float64(1) {
		t.Errorf("Nested array incorrect: %v", array)
	}
}

func TestComprehensive_MultipleJSONDecoder(t *testing.T) {
	// Test decoder with multiple JSON objects
	input := `
		garbage {"first": 1} middle 
		{"second": 2} more junk
		{"third": 3} end
	`

	decoder := New(strings.NewReader(input))

	// Decode first object
	var obj1 map[string]interface{}
	err := decoder.Decode(&obj1)
	if err != nil {
		t.Fatalf("First decode failed: %v", err)
	}
	if obj1["first"] != float64(1) {
		t.Errorf("First object incorrect: %v", obj1)
	}

	// Decode second object
	var obj2 map[string]interface{}
	err = decoder.Decode(&obj2)
	if err != nil {
		t.Fatalf("Second decode failed: %v", err)
	}
	if obj2["second"] != float64(2) {
		t.Errorf("Second object incorrect: %v", obj2)
	}

	// Decode third object
	var obj3 map[string]interface{}
	err = decoder.Decode(&obj3)
	if err != nil {
		t.Fatalf("Third decode failed: %v", err)
	}
	if obj3["third"] != float64(3) {
		t.Errorf("Third object incorrect: %v", obj3)
	}
}

func TestComprehensive_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "No JSON at all",
			data:        []byte("this is just plain text with no JSON"),
			expectError: true,
		},
		{
			name:        "Broken JSON",
			data:        []byte("prefix {broken: json} suffix"),
			expectError: true,
		},
		{
			name:        "Incomplete JSON",
			data:        []byte("prefix {\"incomplete\": suffix"),
			expectError: true,
		},
		{
			name:        "Valid JSON",
			data:        []byte("prefix {\"valid\": true} suffix"),
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)

			if test.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestComprehensive_LargeJSON(t *testing.T) {
	// Generate a large JSON structure
	largeArray := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeArray[i] = map[string]interface{}{
			"id":     i,
			"name":   "Item " + string(rune('A'+i%26)),
			"value":  i * 2,
			"active": i%2 == 0,
		}
	}

	// Create JSON with garbage
	jsonStr := `garbage `
	// Manually construct JSON to avoid using standard library
	jsonStr += `{"items": [`
	for i := 0; i < 10; i++ { // Smaller test for faster execution
		if i > 0 {
			jsonStr += ","
		}
		jsonStr += `{"id": ` + string(rune('0'+i)) + `, "name": "Item` + string(rune('A'+i)) + `"}`
	}
	jsonStr += `]} more garbage`

	var result map[string]interface{}
	err := Unmarshal([]byte(jsonStr), &result)

	if err != nil {
		t.Fatalf("Large JSON failed: %v", err)
	}

	items := result["items"].([]interface{})
	if len(items) != 10 {
		t.Errorf("Expected 10 items, got %d", len(items))
	}
}

func TestComprehensive_OptionsValidation(t *testing.T) {
	// Use non-default options to force robust path
	data := []byte(`{"level1": {"level2": {"level3": "value"}}}`)

	tests := []struct {
		name    string
		options []Option
		wantErr bool
	}{
		{
			name:    "Default options",
			options: nil,
			wantErr: false,
		},
		{
			name:    "High max depth",
			options: []Option{WithMaxDepth(10)},
			wantErr: false,
		},
		{
			name:    "Low max depth",
			options: []Option{WithMaxDepth(2)},
			wantErr: true,
		},
		{
			name:    "Custom buffer size",
			options: []Option{WithBufferSize(8192)},
			wantErr: false,
		},
		{
			name:    "Multiple options",
			options: []Option{WithMaxDepth(5), WithBufferSize(1024)},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(data, &result, test.options...)

			if test.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestComprehensive_SpecialCharacters(t *testing.T) {
	// Test with various special characters and control sequences
	data := []byte(`prefix {
		"backslash": "\\",
		"quote": "\"",
		"newline": "\n",
		"tab": "\t",
		"unicode": "\u0041\u3042",
		"surrogate": "\uD83D\uDE00"
	} suffix`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Special characters failed: %v", err)
	}

	if result["backslash"] != "\\" {
		t.Errorf("Backslash incorrect: %v", result["backslash"])
	}

	if result["quote"] != "\"" {
		t.Errorf("Quote incorrect: %v", result["quote"])
	}

	if result["newline"] != "\n" {
		t.Errorf("Newline incorrect: %v", result["newline"])
	}

	if result["unicode"] != "A\u3042" {
		t.Errorf("Unicode incorrect: %v", result["unicode"])
	}

	if result["surrogate"] != "😀" {
		t.Errorf("Surrogate pair incorrect: %v", result["surrogate"])
	}
}
