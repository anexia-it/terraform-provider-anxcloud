package anxcloud

import "errors"

var (
	// ErrOperationNotSupported is returned when an operation is not supported
	ErrOperationNotSupported = errors.New("operation not supported")
)
