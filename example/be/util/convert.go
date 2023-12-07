package util

import "errors"

func ConvertFromMapTo[T any](vals map[string]any, name string) (*T, error) {
	val, ok := vals[name]
	if !ok {
		return nil, errors.New("value not found")
	}

	casted, ok := val.(T)
	if !ok {
		return nil, errors.New("incorrect type")
	}
	return &casted, nil
}
