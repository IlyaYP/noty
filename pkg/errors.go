package pkg

import "errors"

var (
	ErrNoData          = errors.New("no data")
	ErrInvalidInput    = errors.New("invalid input")
	ErrAlreadyExists   = errors.New("object exists in the DB")
	ErrNotExists       = errors.New("object not exists in the DB")
	ErrServerError     = errors.New("internal server error")
	ErrTooManyRequests = errors.New("too many requests")
)
