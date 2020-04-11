package errors

import "errors"

var (
	NonExistingLocationDeleteAttempt = errors.New("attempting to delete a non-existing location")
	NonExistingLocationUpdateAttempt = errors.New("attempting to update a non-existing location")
	LocationNotFound                 = errors.New("location not found")
)
