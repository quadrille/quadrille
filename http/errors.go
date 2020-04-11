package http

import "errors"

var (
	InvalidBodyErr = errors.New("body should be a valid JSON")
	InvalidDataErr = errors.New("data should be a valid JSON")
)
