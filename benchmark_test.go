package jsonex

import (
	"encoding/json"
	"strings"
	"testing"
)

// Benchmark data sets
var (
	smallJSON = []byte(`{"name": "test", "value": 42, "active": true}`)

	mediumJSON = []byte(`{
		"users": [
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
			{"id": 3, "name": "Charlie", "email": "charlie@example.com"}
		],
		"settings": {
			"theme": "dark",
			"notifications": true,
			"language": "en"
		}
	}`)

	largeJSON = []byte(`{
		"data": [` + strings.Repeat(`{"field1": "value1", "field2": 123, "field3": true},`, 1000) + `
			{"field1": "last", "field2": 999, "field3": false}
		],
		"metadata": {
			"total": 1001,
			"created": "2023-01-01T00:00:00Z",
			"version": "1.0"
		}
	}`)

	deeplyNestedJSON = func() []byte {
		nested := "\"value\""
		for i := 0; i < 50; i++ {
			nested = `{"level` + string(rune('0'+i%10)) + `": ` + nested + `}`
		}
		return []byte(nested)
	}()

	// JSON with garbage prefix/suffix for robust parsing
	robustTestJSON = []byte(`garbage text {"name": "test", "value": 42} more garbage`)
)

// Standard library benchmarks for comparison

func BenchmarkStdLib_Unmarshal_Small(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(smallJSON, &result)
	}
}

func BenchmarkStdLib_Unmarshal_Medium(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(mediumJSON, &result)
	}
}

func BenchmarkStdLib_Unmarshal_Large(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(largeJSON, &result)
	}
}

func BenchmarkStdLib_Unmarshal_DeepNested(b *testing.B) {
	var result interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(deeplyNestedJSON, &result)
	}
}

func BenchmarkStdLib_Decoder_Small(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := json.NewDecoder(strings.NewReader(string(smallJSON)))
		decoder.Decode(&result)
	}
}

func BenchmarkStdLib_Decoder_Medium(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := json.NewDecoder(strings.NewReader(string(mediumJSON)))
		decoder.Decode(&result)
	}
}

// Our implementation benchmarks

func BenchmarkJsonex_Unmarshal_Small(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(smallJSON, &result)
	}
}

func BenchmarkJsonex_Unmarshal_Medium(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(mediumJSON, &result)
	}
}

func BenchmarkJsonex_Unmarshal_Large(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(largeJSON, &result)
	}
}

func BenchmarkJsonex_Unmarshal_DeepNested(b *testing.B) {
	var result interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(deeplyNestedJSON, &result)
	}
}

func BenchmarkJsonex_Unmarshal_Robust(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(robustTestJSON, &result)
	}
}

func BenchmarkJsonex_Decoder_Small(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := New(strings.NewReader(string(smallJSON)))
		decoder.Decode(&result)
	}
}

func BenchmarkJsonex_Decoder_Medium(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := New(strings.NewReader(string(mediumJSON)))
		decoder.Decode(&result)
	}
}

func BenchmarkJsonex_Decoder_Robust(b *testing.B) {
	var result map[string]interface{}
	robustInput := "garbage " + string(smallJSON) + " more garbage"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder := New(strings.NewReader(robustInput))
		decoder.Decode(&result)
	}
}

// Memory allocation benchmarks

func BenchmarkJsonex_Unmarshal_Small_Allocs(b *testing.B) {
	var result map[string]interface{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(smallJSON, &result)
	}
}

func BenchmarkStdLib_Unmarshal_Small_Allocs(b *testing.B) {
	var result map[string]interface{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(smallJSON, &result)
	}
}

// Concurrent benchmarks

func BenchmarkJsonex_Unmarshal_Concurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var result map[string]interface{}
		for pb.Next() {
			Unmarshal(smallJSON, &result)
		}
	})
}

func BenchmarkStdLib_Unmarshal_Concurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var result map[string]interface{}
		for pb.Next() {
			json.Unmarshal(smallJSON, &result)
		}
	})
}

// Options impact benchmarks

func BenchmarkJsonex_Unmarshal_WithOptions(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(smallJSON, &result, WithMaxDepth(100), WithBufferSize(8192))
	}
}

func BenchmarkJsonex_Unmarshal_DefaultOptions(b *testing.B) {
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(smallJSON, &result)
	}
}

// Streaming benchmark - multiple JSON objects

func BenchmarkJsonex_Decoder_MultipleObjects(b *testing.B) {
	input := `{"a":1} garbage {"b":2} more {"c":3}`
	reader := strings.NewReader(input)
	_ = New(reader)

	var result map[string]interface{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader.Reset(input)
		decoder := New(reader)

		// Decode multiple objects
		for range 3 {
			decoder.Decode(&result)
		}
	}
}

// Edge case benchmarks

func BenchmarkJsonex_Unmarshal_EmptyObject(b *testing.B) {
	data := []byte(`{}`)
	var result map[string]interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(data, &result)
	}
}

func BenchmarkJsonex_Unmarshal_EmptyArray(b *testing.B) {
	data := []byte(`[]`)
	var result []interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unmarshal(data, &result)
	}
}
