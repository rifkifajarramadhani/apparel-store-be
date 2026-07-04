package order

import "errors"

var (
	ErrNotFound   = errors.New("order not found")
	ErrEmptyOrder = errors.New("order must contain at least one item")
	ErrOutOfStock = errors.New("insufficient stock")
	ErrInvalidQty = errors.New("quantity must be positive")
)
