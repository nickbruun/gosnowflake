package gosnowflake

import (
	"errors"
)

var (
	// Worker ID out of range.
	ErrWorkerIdOutOfRange = errors.New("worker ID is out of range")

	// Data center ID out of range.
	ErrDatacenterIdOutOfRange = errors.New("data center ID is out of range")

	// Clock is moving backwards.
	ErrClockMovingBackwards = errors.New("system clock is moving backwards")
)
