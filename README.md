# jsonex - Robust JSON Parser for Go

A robust JSON parser library for Go that can extract valid JSON from noisy data streams while maintaining full RFC 8259 compliance.

## Features

- **Robust Parsing**: Extracts the longest valid JSON object or array from data containing garbage/noise
- **RFC 8259 Compliant**: Full compliance with JSON specification including proper escape sequence handling
- **Dual-Mode Operation**: 
  - Fast path for clean JSON using standard library
  - Robust path for extracting JSON from noisy data
- **Streaming Support**: Decoder for processing multiple JSON objects from streams
- **Unicode Support**: Full UTF-8 support including surrogate pairs
- **Configurable**: Customizable depth limits and buffer sizes
- **High Performance**: Competitive performance with Go's standard library

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

Benchmark results show competitive performance with Go's standard library:

```
BenchmarkStdLib_Unmarshal_Small-10      1310816    916.6 ns/op
BenchmarkJsonex_Unmarshal_Small-10      1264669    942.6 ns/op
BenchmarkJsonex_Unmarshal_Robust-10      637006   1874 ns/op
```

The library maintains excellent performance while providing robust parsing capabilities.

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
