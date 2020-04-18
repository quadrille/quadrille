package store

import (
	"errors"
	"fmt"
)

var (
	ErrAddressNotReachable       = errors.New("address not reachable")
	ErrNonExistentLocationDelete = errors.New("cannot delete non existent location")
	ErrNonLeaderNode             = fmt.Errorf("cannot execute operation not leader")
)
