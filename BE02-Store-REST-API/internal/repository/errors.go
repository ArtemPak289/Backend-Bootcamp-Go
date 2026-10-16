package repository

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidInput      = errors.New("invalid input data")
)
