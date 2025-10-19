package storage

import "errors"

var (
	ErrBookingIntervalsCrossed = errors.New("intervals crossed")
	ErrBookingNotFound = errors.New("booking not found")
)