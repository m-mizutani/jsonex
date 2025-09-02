// Package jsonex provides a robust JSON parser that can extract valid JSON objects and arrays
// from input streams that may contain invalid or extraneous data.
//
// The parser is designed to be RFC 8259 compliant and focuses on extracting structured data
// (objects and arrays) while ignoring primitive values that might interfere with robust parsing.
//
// Key features:
// - Extracts JSON objects and arrays from any input stream
// - Skips invalid characters and finds JSON start positions
// - Decoder: extracts the first valid JSON (streaming)
// - Unmarshal: extracts the longest valid JSON (batch processing)
// - Configurable options for depth limits and buffer sizes
// - Comprehensive error reporting with position information
package jsonex

// Re-export key types and functions for convenience

// The main API functions are already defined in their respective files:
// - New() and Decoder.Decode() in decoder.go
// - Unmarshal() in unmarshal.go

// This file serves as the main package documentation and can be extended
// with additional convenience functions or package-level constants if needed.