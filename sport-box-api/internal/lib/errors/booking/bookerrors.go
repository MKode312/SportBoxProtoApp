package bookerrors

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrNotEnoughFundsToPay = errors.New("not enough funds to pay")
	ErrCardNotFound        = errors.New("card not found")
	ErrAlreadyBooked       = errors.New("this box is already booked")
	ErrBookingNotFound     = errors.New("booking not found")
)