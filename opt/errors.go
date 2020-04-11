package opt

import "errors"

var (
	InvalidLatLon = errors.New("invalid lat,lon")
	InvalidData   = errors.New("data must be a valid JSON (without any enclosing quotes)")
)
