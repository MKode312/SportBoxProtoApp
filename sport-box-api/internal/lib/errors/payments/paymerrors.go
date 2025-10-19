package paymerrors

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrNotEnoughFundsToPay = errors.New("not enough funds to pay")
	ErrNotFound            = errors.New("card not found")
)