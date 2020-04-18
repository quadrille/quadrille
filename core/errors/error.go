package errors

import "errors"

var (
	ErrNonExistingLocationDeleteAttempt = errors.New("attempting to delete a non-existing location")
	ErrNonExistingLocationUpdateAttempt = errors.New("attempting to update a non-existing location")
	ErrLocationNotFound                 = errors.New("location not found")
)
