package utils

import (
	"errors"
	"net/url"
	"path/filepath"
)

type StringError string

func (e StringError) Error() string {
	return string(e)
}

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
