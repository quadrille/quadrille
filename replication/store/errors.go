package store

import (
	"errors"
	"fmt"
)

var (
	AddressNotReachableError       = errors.New("address not reachable")
	NonExistentLocationDeleteError = errors.New("cannot delete non existent location")
	NonLeaderNodeError             = fmt.Errorf("cannot execute operation not leader")
)
