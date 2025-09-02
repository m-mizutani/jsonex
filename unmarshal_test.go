package jsonex

import (
	"testing"
)

func TestUnmarshal_BasicObject(t *testing.T) {
	data := []byte(`some garbage {"name": "test", "value": 42} more garbage`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name=test, got %v", result["name"])
	}

	if result["value"] != float64(42) {
		t.Errorf("Expected value=42, got %v", result["value"])
	}
}

func TestUnmarshal_BasicArray(t *testing.T) {
	data := []byte(`garbage [1, 2, 3, "hello"] more stuff`)

	var result []interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(result))
	}

	if result[0] != float64(1) {
		t.Errorf("Expected first element=1, got %v", result[0])
	}

	if result[3] != "hello" {
		t.Errorf("Expected last element=hello, got %v", result[3])
	}
}

func TestUnmarshal_LongestJSON(t *testing.T) {
	// Test that Unmarshal picks the longest valid JSON
	data := []byte(`{"short": 1} prefix {"longer": {"nested": {"deep": "value"}}} suffix`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Should pick the longer JSON
	if _, hasLonger := result["longer"]; !hasLonger {
		t.Errorf("Expected longest JSON to be selected, got %v", result)
	}

	if _, hasShort := result["short"]; hasShort {
		t.Errorf("Shorter JSON should not be selected, got %v", result)
	}
}

func TestUnmarshal_EmptyInput(t *testing.T) {
	data := []byte(``)

	var result interface{}
	err := Unmarshal(data, &result)

	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestUnmarshal_NoValidJSON(t *testing.T) {
	data := []byte(`this is just text with no JSON`)

	var result interface{}
	err := Unmarshal(data, &result)

	if err == nil {
		t.Error("Expected error for input with no valid JSON")
	}
}

func TestUnmarshal_WithOptions(t *testing.T) {
	data := []byte(`{"test": {"nested": {"deep": "value"}}}`)

	var result map[string]interface{}

	// Should work with default max depth
	err := Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal with default options failed: %v", err)
	}

	// Should fail with very low max depth
	err = Unmarshal(data, &result, WithMaxDepth(1))
	if err == nil {
		t.Error("Expected error with max depth 1")
	}
}

func TestUnmarshal_ComplexJSON(t *testing.T) {
	data := []byte(`prefix {
		"users": [
			{"name": "Alice", "age": 30},
			{"name": "Bob", "age": 25}
		],
		"settings": {
			"theme": "dark",
			"notifications": true
		}
	} suffix`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)

	if err != nil {
		t.Fatalf("Unmarshal complex JSON failed: %v", err)
	}

	users, ok := result["users"].([]interface{})
	if !ok {
		t.Fatal("Expected users to be array")
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	settings, ok := result["settings"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected settings to be object")
	}

	if settings["theme"] != "dark" {
		t.Errorf("Expected theme=dark, got %v", settings["theme"])
	}
}