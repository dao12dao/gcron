package common

import "errors"

var (
	ErrorTaskFieldIsNil = errors.New("task field all be required")
	ErrorLockIsOccupied = errors.New("the lock is occupied")
	ErrorNoLocalIPFound = errors.New("no local ip")
)
