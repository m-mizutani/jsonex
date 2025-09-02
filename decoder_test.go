package jsonex

import (
	"strings"
	"testing"
)

func TestDecoder_BasicObject(t *testing.T) {
	input := `garbage {"name": "test", "value": 42} more garbage {"second": true}`
	reader := strings.NewReader(input)

	decoder := New(reader)

	// First decode
	var result1 map[string]interface{}
	err := decoder.Decode(&result1)
	if err != nil {
		t.Fatalf("First Decode failed: %v", err)
	}

	if result1["name"] != "test" {
		t.Errorf("Expected name=test, got %v", result1["name"])
	}

	// Second decode
	var result2 map[string]interface{}
	err = decoder.Decode(&result2)
	if err != nil {
		t.Fatalf("Second Decode failed: %v", err)
	}

	if result2["second"] != true {
		t.Errorf("Expected second=true, got %v", result2["second"])
	}
}

func TestDecoder_Array(t *testing.T) {
	input := `junk [1, 2, 3] more [4, 5]`
	reader := strings.NewReader(input)

	decoder := New(reader)

	// First decode
	var result1 []interface{}
	err := decoder.Decode(&result1)
	if err != nil {
		t.Fatalf("First Decode failed: %v", err)
	}

	if len(result1) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result1))
	}

	// Second decode
	var result2 []interface{}
	err = decoder.Decode(&result2)
	if err != nil {
		t.Fatalf("Second Decode failed: %v", err)
	}

	if len(result2) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(result2))
	}
}

func TestDecoder_WithOptions(t *testing.T) {
	input := `{"deep": {"very": {"nested": "object"}}}`
	reader := strings.NewReader(input)

	// Should fail with low max depth
	decoder := New(reader, WithMaxDepth(2))

	var result map[string]interface{}
	err := decoder.Decode(&result)
	if err == nil {
		t.Error("Expected error with max depth 2")
	}
}

func TestDecoder_MultipleObjects(t *testing.T) {
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

func TestDecoder_ArrayStringVsMap(t *testing.T) {
	// Test decoder behavior with arrays vs maps containing array-like strings
	// Note: Decoder uses parseNext() which finds FIRST valid JSON, not longest
	tests := []struct {
		name            string
		input           string
		expectFirstType string // "array" or "map"
		description     string
	}{
		{
			name:            "Array first, then map with array string",
			input:           `[1] noise {"array": "[1]"} garbage`,
			expectFirstType: "array",
			description:     "Decoder finds [1] first, ignoring later map",
		},
		{
			name:            "Map first, then map with object string",
			input:           `{"fake": "json"} text {"real": "{\"fake\": \"json\"}"} end`,
			expectFirstType: "map",
			description:     "Decoder finds first map, not considering content",
		},
		{
			name:            "Map first with array string content",
			input:           `noise {"array": "[1]"} then [1] garbage`,
			expectFirstType: "map",
			description:     "Map appears first, so it's selected",
		},
		{
			name:            "Array first, then more complex map",
			input:           `[1, 2] [3, 4] {"nums": "[1, 2, 3]"} [5]`,
			expectFirstType: "array",
			description:     "First array [1, 2] is selected",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoder := New(strings.NewReader(test.input))

			if test.expectFirstType == "map" {
				var result map[string]interface{}
				err := decoder.Decode(&result)
				if err != nil {
					t.Fatalf("Map decode failed (%s): %v", test.description, err)
				}

				if len(result) == 0 {
					t.Error("Result map should not be empty")
				}
				t.Logf("Successfully decoded map: %v", result)
			} else {
				var result []interface{}
				err := decoder.Decode(&result)
				if err != nil {
					t.Fatalf("Array decode failed (%s): %v", test.description, err)
				}

				if len(result) == 0 {
					t.Error("Result array should not be empty")
				}
				t.Logf("Successfully decoded array: %v", result)
			}
		})
	}
}

func TestDecoder_FirstVsLongestJSON(t *testing.T) {
	// Test demonstrating difference between Decoder (first) and Unmarshal (longest)
	input := `[1] {"large": {"nested": {"structure": "with more content"}}, "multiple": "fields"} [2]`

	// Test Decoder - should get first JSON ([1])
	t.Run("Decoder gets first JSON", func(t *testing.T) {
		decoder := New(strings.NewReader(input))
		
		var arrayResult []interface{}
		err := decoder.Decode(&arrayResult)
		if err != nil {
			t.Fatalf("Decoder failed: %v", err)
		}
		
		if len(arrayResult) != 1 || arrayResult[0] != float64(1) {
			t.Errorf("Expected [1], got %v", arrayResult)
		}
	})

	// Test Unmarshal - should get longest JSON (the map)
	t.Run("Unmarshal gets longest JSON", func(t *testing.T) {
		var mapResult map[string]interface{}
		err := Unmarshal([]byte(input), &mapResult)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		
		if _, ok := mapResult["large"]; !ok {
			t.Errorf("Expected longest JSON (map with 'large' key), got %v", mapResult)
		}
		
		if len(mapResult) != 2 { // Should have "large" and "multiple" keys
			t.Errorf("Expected 2 keys in longest JSON, got %d: %v", len(mapResult), mapResult)
		}
	})
}

