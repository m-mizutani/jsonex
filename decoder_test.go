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