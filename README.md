# jsonex - Go JSON Extractor for Unclean Text

[![Unit test](https://github.com/m-mizutani/jsonex/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/jsonex/actions/workflows/test.yml)
[![Lint](https://github.com/m-mizutani/jsonex/actions/workflows/lint.yml/badge.svg)](https://github.com/m-mizutani/jsonex/actions/workflows/lint.yml)
[![Gosec](https://github.com/m-mizutani/jsonex/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/jsonex/actions/workflows/gosec.yml)
[![trivy](https://github.com/m-mizutani/jsonex/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/jsonex/actions/workflows/trivy.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/jsonex.svg)](https://pkg.go.dev/github.com/m-mizutani/jsonex)

A JSON parser library for Go that can extract valid JSON from noisy data streams while maintaining RFC 8259 compliance.

## Features

- **Parsing from noisy data**: Extracts the longest valid JSON object or array from data containing garbage/noise
- **RFC 8259 Compliant**: Full compliance with JSON specification including proper escape sequence handling
- **Dual-Mode Operation**: 
  - Fast path for clean JSON using standard library
  - Fallback path for extracting JSON from noisy data
- **Streaming Support**: Decoder for processing multiple JSON objects from streams
- **Unicode Support**: Full UTF-8 support including surrogate pairs
- **Configurable**: Customizable depth limits and buffer sizes

## Installation

```bash
go get github.com/m-mizutani/jsonex
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/m-mizutani/jsonex"
)

func main() {
    // Extract JSON from noisy data
    data := []byte(`garbage {"name": "John", "age": 30} more noise`)
    
    var result map[string]interface{}
    err := jsonex.Unmarshal(data, &result)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result["name"]) // Output: John
    fmt.Println(result["age"])  // Output: 30
}
```

### Streaming Decoder

```go
package main

import (
    "fmt"
    "strings"
    "github.com/m-mizutani/jsonex"
)

func main() {
    input := `noise {"first": 1} garbage {"second": 2} end`
    decoder := jsonex.New(strings.NewReader(input))
    
    var obj1, obj2 map[string]interface{}
    
    // Decode first JSON object
    if err := decoder.Decode(&obj1); err != nil {
        panic(err)
    }
    
    // Decode second JSON object  
    if err := decoder.Decode(&obj2); err != nil {
        panic(err)
    }
    
    fmt.Println(obj1["first"])  // Output: 1
    fmt.Println(obj2["second"]) // Output: 2
}
```

### With Options

```go
package main

import (
    "github.com/m-mizutani/jsonex"
)

func main() {
    data := []byte(`{"deeply": {"nested": {"json": "value"}}}`)
    
    var result map[string]interface{}
    err := jsonex.Unmarshal(data, &result,
        jsonex.WithMaxDepth(10),      // Set maximum nesting depth
        jsonex.WithBufferSize(8192),  // Set buffer size
    )
    if err != nil {
        panic(err)
    }
}
```

## API Reference

### Functions

#### `Unmarshal(data []byte, v interface{}, opts ...Option) error`

Parses JSON-encoded data and stores the result in the value pointed to by v. Unlike standard `json.Unmarshal`, this function extracts the longest valid JSON object or array from the input data, ignoring any preceding or trailing invalid content.

#### `New(r io.Reader, opts ...Option) *Decoder`

Creates a new Decoder that reads from r.

### Types

#### `Decoder`

```go
type Decoder struct {
    // contains filtered or unexported fields
}

func (d *Decoder) Decode(v interface{}) error
```

### Options

#### `WithMaxDepth(depth int) Option`

Sets the maximum nesting depth for JSON parsing (default: 1000).

#### `WithBufferSize(size int) Option`  

Sets the buffer size for internal operations (default: 4096).

## RFC 8259 Compliance

This library is fully compliant with RFC 8259 (The JavaScript Object Notation Data Interchange Format):

- ✅ Proper JSON grammar support (objects, arrays, strings, numbers, booleans, null)
- ✅ Complete escape sequence handling (`\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`)
- ✅ Unicode escape sequences (`\uXXXX`) including surrogate pairs
- ✅ UTF-8 character encoding support
- ✅ Syntax validation and error reporting
- ✅ Whitespace handling
- ✅ Number format validation

## Performance

Benchmark results:

```
BenchmarkStdLib_Unmarshal_Small-10      1310816    916.6 ns/op
BenchmarkJsonex_Unmarshal_Small-10      1264669    942.6 ns/op
BenchmarkJsonex_Unmarshal_Robust-10      637006   1874 ns/op
```

Note: The robust parsing mode has additional overhead compared to the standard library due to the extra processing required for handling noisy data.

## Error Handling

The library provides detailed error information including:

- Error type classification (syntax, unicode, escape, EOF, invalid JSON)
- Position information (line, column, offset)
- Contextual error messages

```go
if err := jsonex.Unmarshal(data, &result); err != nil {
    if jsonErr, ok := err.(*jsonex.Error); ok {
        fmt.Printf("Error at line %d, column %d: %s\n", 
            jsonErr.Position.Line, 
            jsonErr.Position.Column, 
            jsonErr.Message)
    }
}
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Fuzzing

This library includes fuzzing tests to validate behavior against various inputs.

### Running Fuzzing Tests

#### Using Task Runner (Recommended)

```bash
# Install Task runner first: https://taskfile.dev/installation/

# Run all fuzzing tests
task fuzz

# Run individual fuzzing tests
task fuzz:unmarshal      # Test Unmarshal function
task fuzz:decoder        # Test Decoder function  
task fuzz:unicode        # Test Unicode handling
task fuzz:deep-nesting   # Test deep nesting structures

# Run extended fuzzing (5 minutes each)
task fuzz:long

# Clean fuzzing corpus and cache
task fuzz:clean
```

#### Using Go Commands Directly

```bash
# Run individual fuzz tests
go test -fuzz=FuzzUnmarshal -fuzztime=30s
go test -fuzz=FuzzDecoder -fuzztime=30s
go test -fuzz=FuzzUnicodeHandling -fuzztime=30s
go test -fuzz=FuzzDeepNesting -fuzztime=30s

# Run for longer duration
go test -fuzz=FuzzUnmarshal -fuzztime=5m
```

### Fuzzing Coverage

The fuzzing tests cover:

- **Unmarshal Function**: Various JSON inputs including malformed data, edge cases, and robust parsing scenarios
- **Decoder Function**: Streaming JSON processing with multiple objects and incomplete data
- **Unicode Handling**: UTF-8 validation, escape sequences, surrogate pairs, and invalid sequences
- **Deep Nesting**: Extremely nested structures to test stack limits and memory usage

### Fuzzing Corpus

Fuzzing corpus files are stored in `testdata/fuzz/` and include automatically generated test cases that increase code coverage. The corpus is managed automatically but can be cleaned with `task fuzz:clean`.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
