package storage

import "errors"

var (
	ErrBookingIntervalsCrossed = errors.New("intervals crossed")
	ErrBookingNotFound = errors.New("booking not found")
	ErrAlreadyBooked = errors.New("this box is already booked")
	ErrNotYourBooking = errors.New("this booking belongs to another user")
)