package storage

import "errors"

var (
	ErrCardExists = errors.New("card already exists")
	ErrCardNotFound = errors.New("card not found")
	ErrNotEnoughFundsToPay = errors.New("not enough funds to pay")
)