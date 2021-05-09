package tyr

import "errors"

// package errors
var (
	ErrNotFound           = errors.New("not found")
	ErrNotSupported       = errors.New("not supported")
	ErrTableNotSpecified  = errors.New("table not specified")
	ErrColumnNotSpecified = errors.New("column not specified")
	ErrInvalidPointer     = errors.New("attempt to load into an invalid pointer")
	ErrPlaceholderCount   = errors.New("wrong placeholder count")
	ErrInvalidSliceLength = errors.New("length of slice is 0. length must be >= 1")
	ErrCantConvertToTime  = errors.New("can't convert to time.Time")
	ErrInvalidTimestring  = errors.New("invalid time string")
)
