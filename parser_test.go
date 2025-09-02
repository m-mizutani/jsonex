package jsonex

import (
	"testing"
)

func TestParser_MalformedJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		wantErr bool
	}{
		{
			name:    "Missing closing brace",
			data:    []byte(`{"key": "value"`),
			wantErr: true,
		},
		{
			name:    "Missing closing bracket",
			data:    []byte(`[1, 2, 3`),
			wantErr: true,
		},
		{
			name:    "Extra comma object",
			data:    []byte(`{"key": "value",}`),
			wantErr: true,
		},
		{
			name:    "Extra comma array",
			data:    []byte(`[1, 2, 3,]`),
			wantErr: true,
		},
		{
			name:    "Missing quotes on key",
			data:    []byte(`{key: "value"}`),
			wantErr: true,
		},
		{
			name:    "Single quotes",
			data:    []byte(`{'key': 'value'}`),
			wantErr: true,
		},
		{
			name:    "Unescaped quotes in string",
			data:    []byte(`{"key": "value"with"quotes"}`),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			
			if test.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !test.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParser_MaxDepthEnforcement(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		shouldErr bool
	}{
		{
			name:      "Depth 5 - should fail",
			depth:     5,
			shouldErr: true,
		},
		{
			name:      "Depth 8 - should fail", 
			depth:     8,
			shouldErr: true,
		},
		{
			name:      "Depth 12 - should pass",
			depth:     12,
			shouldErr: false,
		},
		{
			name:      "Depth 15 - should pass",
			depth:     15,
			shouldErr: false,
		},
	}

	// Create deeply nested JSON (10 levels)
	deepJSON := `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":"value"}}}}}}}}}`

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal([]byte(deepJSON), &result, WithMaxDepth(test.depth))
			
			if test.shouldErr && err == nil {
				t.Error("Expected depth error but got none")
			}
			if !test.shouldErr && err != nil {
				t.Errorf("Unexpected depth error: %v", err)
			}
		})
	}
}

func TestParser_NumberFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Integer",
			data: []byte(`{"num": 42}`),
		},
		{
			name: "Float",
			data: []byte(`{"num": 3.14}`),
		},
		{
			name: "Scientific notation",
			data: []byte(`{"num": 1.5e10}`),
		},
		{
			name: "Negative scientific",
			data: []byte(`{"num": -2.5E-5}`),
		},
		{
			name: "Zero",
			data: []byte(`{"num": 0}`),
		},
		{
			name: "Negative zero",
			data: []byte(`{"num": -0}`),
		},
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

func TestParser_NestedStructures(t *testing.T) {
	// Complex nested structure
	data := []byte(`{
		"level1": {
			"array": [
				{"nested": true},
				[1, 2, {"deep": "value"}]
			],
			"level2": {
				"more": "data"
			}
		}
	}`)

	var result map[string]interface{}
	err := Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Nested structure parsing failed: %v", err)
	}

	// Basic validation
	level1, ok := result["level1"].(map[string]interface{})
	if !ok {
		t.Error("level1 is not an object")
	}

	array, ok := level1["array"].([]interface{})
	if !ok {
		t.Error("array is not an array")
	}

	if len(array) != 2 {
		t.Errorf("Expected array length 2, got %d", len(array))
	}
}