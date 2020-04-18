package http

import "errors"

var (
	ErrInvalidBody           = errors.New("body should be a valid JSON")
	ErrInvalidData           = errors.New("data should be a valid JSON")
	ErrInvalidBulkWriteArray = errors.New("body should contain an array of insert/update operations")
)
