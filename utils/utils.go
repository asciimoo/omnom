// Package utils provides utility functions and types used across the Omnom application.
//
// This package contains helper functions for common operations including:
//   - String-based error types for constant errors
//   - Key-value data structure construction from variadic arguments
//   - File extension extraction from URLs and paths
//
// These utilities are used throughout the codebase to reduce code duplication
// and provide consistent behavior for common operations.
//
// Example usage:
//
//	// Create a string error
//	const ErrNotFound = utils.StringError("not found")
//
//	// Build a map from key-value pairs
//	data, err := utils.KVData("name", "John", "age", 30)
//
//	// Extract file extension
//	ext := utils.GetExtension("https://example.com/image.jpg")
package utils

import (
	"errors"
	"net/url"
	"path/filepath"
)

// StringError is a string that implements the error interface.
type StringError string

func (e StringError) Error() string {
	return string(e)
}

// KVData creates a map from alternating key-value arguments.
func KVData(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// GetExtension extracts the file extension from a URL or path.
func GetExtension(s string) string {
	defaultExt := ".ext"
	pu, err := url.Parse(s)
	if err != nil {
		return defaultExt
	}
	ext := filepath.Ext(pu.Path)
	if ext == "" {
		return defaultExt
	}
	return ext
}
