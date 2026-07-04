package catalog

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid catalog input")
	ErrConflict     = errors.New("resource already exists")
	ErrReferenced   = errors.New("resource is still referenced")
)
