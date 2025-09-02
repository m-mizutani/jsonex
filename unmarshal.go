package jsonex

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v
// Unlike the standard json.Unmarshal, this function extracts the longest valid JSON 
// object or array from the input data, ignoring any preceding or trailing invalid content
func Unmarshal(data []byte, v interface{}, opts ...Option) error {
	if len(data) == 0 {
		return newInvalidJSONError(position{}, "empty input data")
	}

	options := applyOptions(opts...)

	// Fast path: try standard library first if data looks clean and no special options
	if options.maxDepth == 1000 && options.bufferSize == 4096 { // Default options only
		trimmed := bytes.TrimSpace(data)
		if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
			// Check if the trimmed data equals the original data (no garbage)
			if bytes.Equal(trimmed, data) {
				if err := json.Unmarshal(trimmed, v); err == nil {
					return nil
				}
			}
		}
	}

	// Robust path: find and extract the longest valid JSON
	jsonBytes, err := parseLongest(data, options)
	if err != nil {
		return err
	}

	// Use standard library to decode the extracted JSON
	err = json.Unmarshal(jsonBytes, v)
	if err != nil {
		return err
	}
	
	// Post-process to decode escape sequences for RFC compliance
	processEscapeSequences(v)
	return nil
}

// processEscapeSequences recursively processes the value to decode escape sequences
func processEscapeSequences(v interface{}) {
	if v == nil {
		return
	}
	
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	switch val.Kind() {
	case reflect.Map:
		for _, key := range val.MapKeys() {
			mapVal := val.MapIndex(key)
			if mapVal.Kind() == reflect.Interface {
				interfaceVal := mapVal.Interface()
				if str, ok := interfaceVal.(string); ok {
					// Decode escape sequences in string
					decoded := decodeEscapeSequences(str)
					val.SetMapIndex(key, reflect.ValueOf(decoded))
				} else {
					// Recursively process nested values
					processEscapeSequences(interfaceVal)
				}
			}
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() == reflect.Interface {
				interfaceVal := elem.Interface()
				if str, ok := interfaceVal.(string); ok {
					// Decode escape sequences in string
					decoded := decodeEscapeSequences(str)
					elem.Set(reflect.ValueOf(decoded))
				} else {
					// Recursively process nested values
					processEscapeSequences(interfaceVal)
				}
			}
		}
	}
}

// decodeEscapeSequences decodes JSON escape sequences in a string
func decodeEscapeSequences(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if i < len(s)-1 && s[i] == '\\' {
			switch s[i+1] {
			case '"':
				result.WriteByte('"')
				i += 2
			case '\\':
				result.WriteByte('\\')
				i += 2
			case '/':
				result.WriteByte('/')
				i += 2
			case 'b':
				result.WriteByte('\b')
				i += 2
			case 'f':
				result.WriteByte('\f')
				i += 2
			case 'n':
				result.WriteByte('\n')
				i += 2
			case 'r':
				result.WriteByte('\r')
				i += 2
			case 't':
				result.WriteByte('\t')
				i += 2
			case 'u':
				if i+5 < len(s) {
					hexStr := s[i+2 : i+6]
					if codePoint, err := strconv.ParseUint(hexStr, 16, 16); err == nil {
						// Handle surrogate pairs
						if codePoint >= 0xD800 && codePoint <= 0xDBFF && i+11 < len(s) && s[i+6:i+8] == "\\u" {
							// High surrogate followed by potential low surrogate
							lowHexStr := s[i+8 : i+12]
							if lowCodePoint, err := strconv.ParseUint(lowHexStr, 16, 16); err == nil && lowCodePoint >= 0xDC00 && lowCodePoint <= 0xDFFF {
								// Valid surrogate pair
								runeValue := 0x10000 + ((codePoint&0x3FF)<<10) + (lowCodePoint&0x3FF)
								result.WriteRune(rune(runeValue))
								i += 12
								continue
							}
						}
						result.WriteRune(rune(codePoint))
						i += 6
					} else {
						result.WriteByte(s[i])
						i++
					}
				} else {
					result.WriteByte(s[i])
					i++
				}
			default:
				result.WriteByte(s[i])
				i++
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}