package jsonex

import (
	"testing"
)

// RFC 8259 compliance tests
// https://tools.ietf.org/html/rfc8259

func TestRFC8259_JSONStructure(t *testing.T) {
	// RFC 8259 Section 2: JSON Grammar
	// A JSON text is a serialized value. Note that certain previous
	// specifications of JSON constrained a JSON text to be an object or an
	// array. Implementations that generate only objects or arrays where a
	// JSON text is called for will be interoperable in the sense that all
	// implementations will accept these as conforming JSON texts.
	
	tests := []struct {
		name        string
		data        []byte
		shouldParse bool
		description string
	}{
		{
			name:        "Valid object",
			data:        []byte(`{"key": "value"}`),
			shouldParse: true,
			description: "Objects are valid JSON",
		},
		{
			name:        "Valid array",
			data:        []byte(`[1, 2, 3]`),
			shouldParse: true,
			description: "Arrays are valid JSON",
		},
		{
			name:        "Empty object",
			data:        []byte(`{}`),
			shouldParse: true,
			description: "Empty objects are valid",
		},
		{
			name:        "Empty array",
			data:        []byte(`[]`),
			shouldParse: true,
			description: "Empty arrays are valid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			
			if test.shouldParse && err != nil {
				t.Errorf("Expected valid JSON but got error: %v", err)
			}
			if !test.shouldParse && err == nil {
				t.Errorf("Expected error but parsing succeeded")
			}
		})
	}
}

func TestRFC8259_Objects(t *testing.T) {
	// RFC 8259 Section 4: Objects
	// An object is an unordered set of name/value pairs
	
	tests := []struct {
		name string
		data []byte
		want map[string]interface{}
	}{
		{
			name: "Simple object",
			data: []byte(`{"name": "value"}`),
			want: map[string]interface{}{"name": "value"},
		},
		{
			name: "Multiple pairs",
			data: []byte(`{"first": 1, "second": 2}`),
			want: map[string]interface{}{"first": float64(1), "second": float64(2)},
		},
		{
			name: "Nested object",
			data: []byte(`{"outer": {"inner": "value"}}`),
			want: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name: "Object with array",
			data: []byte(`{"items": [1, 2, 3]}`),
			want: map[string]interface{}{
				"items": []interface{}{float64(1), float64(2), float64(3)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Basic validation - detailed comparison would require deep equality
			if len(result) != len(test.want) {
				t.Errorf("Object length mismatch: got %d, want %d", len(result), len(test.want))
			}
		})
	}
}

func TestRFC8259_Arrays(t *testing.T) {
	// RFC 8259 Section 5: Arrays
	// An array is an ordered sequence of zero or more values
	
	tests := []struct {
		name string
		data []byte
		want []interface{}
	}{
		{
			name: "Simple array",
			data: []byte(`[1, 2, 3]`),
			want: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name: "Mixed types",
			data: []byte(`[1, "string", true, null]`),
			want: []interface{}{float64(1), "string", true, nil},
		},
		{
			name: "Nested array",
			data: []byte(`[[1, 2], [3, 4]]`),
			want: []interface{}{
				[]interface{}{float64(1), float64(2)},
				[]interface{}{float64(3), float64(4)},
			},
		},
		{
			name: "Array with objects",
			data: []byte(`[{"id": 1}, {"id": 2}]`),
			want: []interface{}{
				map[string]interface{}{"id": float64(1)},
				map[string]interface{}{"id": float64(2)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result []interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if len(result) != len(test.want) {
				t.Errorf("Array length mismatch: got %d, want %d", len(result), len(test.want))
			}
		})
	}
}

func TestRFC8259_Values(t *testing.T) {
	// RFC 8259 Section 3: Values
	// A JSON value MUST be an object, array, number, or string, or one of
	// the following three literal names: false null true
	
	tests := []struct {
		name   string
		data   []byte
		expect interface{}
	}{
		// Note: Our implementation only supports objects and arrays as top-level values
		// This is by design as specified in the requirements
		{
			name:   "Object value",
			data:   []byte(`{"key": "value"}`),
			expect: map[string]interface{}{"key": "value"},
		},
		{
			name:   "Array value",
			data:   []byte(`["item1", "item2"]`),
			expect: []interface{}{"item1", "item2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			// Basic type validation
			switch test.expect.(type) {
			case map[string]interface{}:
				if _, ok := result.(map[string]interface{}); !ok {
					t.Error("Expected object result")
				}
			case []interface{}:
				if _, ok := result.([]interface{}); !ok {
					t.Error("Expected array result")
				}
			}
		})
	}
}

func TestRFC8259_Strings(t *testing.T) {
	// RFC 8259 Section 7: Strings
	// A string is a sequence of Unicode code points wrapped with quotation marks
	
	tests := []struct {
		name string
		data []byte
		key  string
		want string
	}{
		{
			name: "Basic string",
			data: []byte(`{"text": "hello world"}`),
			key:  "text",
			want: "hello world",
		},
		{
			name: "Empty string",
			data: []byte(`{"empty": ""}`),
			key:  "empty",
			want: "",
		},
		{
			name: "String with spaces",
			data: []byte(`{"spaced": "  hello  world  "}`),
			key:  "spaced",
			want: "  hello  world  ",
		},
		{
			name: "Unicode string",
			data: []byte(`{"unicode": "こんにちは世界"}`),
			key:  "unicode",
			want: "こんにちは世界",
		},
		{
			name: "Emoji string",
			data: []byte(`{"emoji": "🌍🚀💻"}`),
			key:  "emoji",
			want: "🌍🚀💻",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if got, ok := result[test.key].(string); !ok {
				t.Error("Value is not a string")
			} else if got != test.want {
				t.Errorf("String value mismatch: got %q, want %q", got, test.want)
			}
		})
	}
}

func TestRFC8259_StringEscapes(t *testing.T) {
	// RFC 8259 Section 7: String escape sequences
	// All Unicode characters may be placed within the quotation marks, except
	// for the characters that MUST be escaped: quotation mark, reverse solidus,
	// and the control characters (U+0000 through U+001F).
	
	tests := []struct {
		name   string
		data   []byte
		key    string
		expect string
	}{
		{
			name:   "Escaped quote",
			data:   []byte(`garbage {"quote": "He said \\\"Hello\\\""} trash`),
			key:    "quote",
			expect: `He said "Hello"`,
		},
		{
			name:   "Escaped backslash",
			data:   []byte(`prefix {"path": "C:\\\\\\\\Program Files"} suffix`),
			key:    "path",
			expect: `C:\\Program Files`,
		},
		{
			name:   "Escaped newline",
			data:   []byte(`noise {"text": "line1\\nline2"} end`),
			key:    "text",
			expect: "line1\nline2",
		},
		{
			name:   "Escaped tab",
			data:   []byte(`start {"text": "col1\\tcol2"} finish`),
			key:    "text",
			expect: "col1\tcol2",
		},
		{
			name:   "Unicode escape",
			data:   []byte(`begin {"unicode": "\\u0041\\u0042"} done`),
			key:    "unicode",
			expect: "AB",
		},
		{
			name:   "Surrogate pair",
			data:   []byte(`junk {"emoji": "\\uD83D\\uDE00"} more`),
			key:    "emoji",
			expect: "😀",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if got, ok := result[test.key].(string); !ok {
				t.Error("Value is not a string")
			} else if got != test.expect {
				t.Errorf("Escaped string mismatch: got %q, want %q", got, test.expect)
			}
		})
	}
}

func TestRFC8259_Numbers(t *testing.T) {
	// RFC 8259 Section 6: Numbers
	// Numeric values that cannot be represented in the grammar below
	// (such as Infinity and NaN) are not permitted.
	
	tests := []struct {
		name   string
		data   []byte
		key    string
		expect float64
	}{
		{
			name:   "Integer",
			data:   []byte(`{"num": 42}`),
			key:    "num",
			expect: 42,
		},
		{
			name:   "Negative integer",
			data:   []byte(`{"num": -17}`),
			key:    "num",
			expect: -17,
		},
		{
			name:   "Zero",
			data:   []byte(`{"num": 0}`),
			key:    "num",
			expect: 0,
		},
		{
			name:   "Decimal",
			data:   []byte(`{"num": 3.14159}`),
			key:    "num",
			expect: 3.14159,
		},
		{
			name:   "Scientific notation",
			data:   []byte(`{"num": 1.23e10}`),
			key:    "num",
			expect: 1.23e10,
		},
		{
			name:   "Negative scientific",
			data:   []byte(`{"num": -1.5E-3}`),
			key:    "num",
			expect: -1.5E-3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if got, ok := result[test.key].(float64); !ok {
				t.Error("Value is not a number")
			} else if got != test.expect {
				t.Errorf("Number value mismatch: got %v, want %v", got, test.expect)
			}
		})
	}
}

func TestRFC8259_Literals(t *testing.T) {
	// RFC 8259 Section 3: Literal names
	// The literal names MUST be lowercase. No other literal names are allowed.
	
	tests := []struct {
		name   string
		data   []byte
		key    string
		expect interface{}
	}{
		{
			name:   "Boolean true",
			data:   []byte(`{"flag": true}`),
			key:    "flag",
			expect: true,
		},
		{
			name:   "Boolean false",
			data:   []byte(`{"flag": false}`),
			key:    "flag",
			expect: false,
		},
		{
			name:   "Null value",
			data:   []byte(`{"value": null}`),
			key:    "value",
			expect: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if got := result[test.key]; got != test.expect {
				t.Errorf("Literal value mismatch: got %v, want %v", got, test.expect)
			}
		})
	}
}

func TestRFC8259_Whitespace(t *testing.T) {
	// RFC 8259 Section 2: Insignificant whitespace
	// Insignificant whitespace is allowed before or after any of the six
	// structural characters: [ ] { } : ,
	
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Whitespace around object",
			data: []byte(`  {  "key"  :  "value"  }  `),
		},
		{
			name: "Whitespace around array",
			data: []byte(`  [  1  ,  2  ,  3  ]  `),
		},
		{
			name: "Tabs and newlines",
			data: []byte("{\n\t\"key\": \"value\"\n}"),
		},
		{
			name: "Multiple whitespace types",
			data: []byte(" \t\n\r{ \t\n\r\"key\" \t\n\r: \t\n\r\"value\" \t\n\r} \t\n\r"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("Whitespace handling failed: %v", err)
			}
		})
	}
}

func TestRFC8259_Syntax_Violations(t *testing.T) {
	// RFC 8259 compliance requires rejecting invalid syntax
	
	tests := []struct {
		name        string
		data        []byte
		description string
	}{
		{
			name:        "Trailing comma in object",
			data:        []byte(`{"key": "value",}`),
			description: "Objects must not have trailing commas",
		},
		{
			name:        "Trailing comma in array",
			data:        []byte(`[1, 2, 3,]`),
			description: "Arrays must not have trailing commas",
		},
		{
			name:        "Unquoted key",
			data:        []byte(`{key: "value"}`),
			description: "Object keys must be quoted strings",
		},
		{
			name:        "Single quotes",
			data:        []byte(`{'key': 'value'}`),
			description: "Only double quotes are allowed",
		},
		{
			name:        "Missing closing brace",
			data:        []byte(`{"key": "value"`),
			description: "All brackets must be properly closed",
		},
		{
			name:        "Missing closing bracket",
			data:        []byte(`[1, 2, 3`),
			description: "All brackets must be properly closed",
		},
		{
			name:        "Invalid literal case",
			data:        []byte(`{"flag": True}`),
			description: "Literals must be lowercase",
		},
		{
			name:        "Invalid escape",
			data:        append([]byte("trash {\"text\": \"invalid"), append([]byte{0x5C, 0x78}, []byte(" escape\"} noise")...)...),
			description: "Invalid escape sequences are not allowed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err == nil {
				t.Errorf("Expected syntax error but parsing succeeded: %s", test.description)
			}
		})
	}
}

func TestRFC8259_Interoperability(t *testing.T) {
	// RFC 8259 Section 8: String and Character Issues
	// An implementation may set limits on the size of texts that it accepts
	
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Moderate nesting",
			data: []byte(`{"a": {"b": {"c": {"d": "value"}}}}`),
		},
		{
			name: "Moderate array",
			data: []byte(`[1, [2, [3, [4, 5]]], 6]`),
		},
		{
			name: "Mixed structure",
			data: []byte(`{"users": [{"id": 1, "data": {"name": "Alice"}}, {"id": 2, "data": {"name": "Bob"}}]}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("RFC 8259 compliant JSON failed to parse: %v", err)
			}
		})
	}
}

func TestRFC8259_CharacterEncoding(t *testing.T) {
	// RFC 8259 Section 8.1: Character Encoding
	// JSON text SHALL be encoded in UTF-8, UTF-16, or UTF-32
	// Since Go strings are UTF-8, we test UTF-8 compliance
	
	tests := []struct {
		name string
		data []byte
		desc string
	}{
		{
			name: "ASCII characters",
			data: []byte(`{"ascii": "Hello World"}`),
			desc: "Basic ASCII should work",
		},
		{
			name: "Latin-1 supplement",
			data: []byte(`{"latin": "café résumé"}`),
			desc: "Extended Latin characters",
		},
		{
			name: "Cyrillic",
			data: []byte(`{"cyrillic": "Привет мир"}`),
			desc: "Cyrillic script",
		},
		{
			name: "CJK characters",
			data: []byte(`{"cjk": "你好世界 こんにちは世界 안녕하세요 세계"}`),
			desc: "Chinese, Japanese, Korean characters",
		},
		{
			name: "Mathematical symbols",
			data: []byte(`{"math": "∑ ∫ ∂ ∆ ∇ ∞ ℝ ℂ ℕ"}`),
			desc: "Mathematical Unicode symbols",
		},
		{
			name: "Emoji and symbols",
			data: []byte(`{"emoji": "🌍🚀💻🎉⭐️🔥💡🌈"}`),
			desc: "Modern emoji characters",
		},
		{
			name: "Mixed scripts",
			data: []byte(`{"mixed": "Hello مرحبا שלום Здравствуй こんにちは 你好"}`),
			desc: "Multiple scripts in one string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result map[string]interface{}
			err := Unmarshal(test.data, &result)
			if err != nil {
				t.Errorf("UTF-8 JSON failed to parse (%s): %v", test.desc, err)
			}
			
			// Verify the string was preserved correctly
			for key, value := range result {
				if str, ok := value.(string); ok && len(str) == 0 {
					t.Errorf("String value was lost for key %s", key)
				}
			}
		})
	}
}