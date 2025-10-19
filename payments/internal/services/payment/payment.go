package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"payments/internal/lib/logger/sl"
	"payments/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type Payment struct {
	log             *slog.Logger
	cardAdder       CardAdder
	fundsAdder      FundsAdder
	paymentProvider PaymentProvider
}

type CardAdder interface {
	AddCard(ctx context.Context, email string, cardNumber []byte, cvc []byte, phoneNumber []byte) (success bool, err error)
}

type FundsAdder interface {
	AddFunds(ctx context.Context, email string, amount int64) (balance int64, success bool, err error)
}

type PaymentProvider interface {
	Pay(ctx context.Context, email string, amount int64) (balance int64, success bool, err error)
}

const (
	emptyBalanceValue = -1
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrNotEnoughFundsToPay = errors.New("not enough funds to pay")
	ErrNotFound            = errors.New("card not found")
)

func New(log *slog.Logger, cardAdder CardAdder, fundsAdder FundsAdder, paymentProvider PaymentProvider) *Payment {
	return &Payment{
		log:             log,
		cardAdder:       cardAdder,
		fundsAdder:      fundsAdder,
		paymentProvider: paymentProvider,
	}
}

func (p *Payment) AddCard(ctx context.Context, email string, cardNumber string, cvc string, phoneNumber string) (bool, error) {
	const op = "payment.AddCard"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("adding a card")

	cardNumberHash, err := bcrypt.GenerateFromPassword([]byte(cardNumber), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate card number hash", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	cvcHash, err := bcrypt.GenerateFromPassword([]byte(cvc), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate cvc hash", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	phoneNumberHash, err := bcrypt.GenerateFromPassword([]byte(phoneNumber), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate phone number hash", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	success, err := p.cardAdder.AddCard(ctx, email, cardNumberHash, cvcHash, phoneNumberHash)
	if err != nil {
		if errors.Is(err, storage.ErrCardExists) {
			log.Error("card already exists", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to save a card", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("card added")

	return success, nil
}

func (p *Payment) AddFunds(ctx context.Context, email string, amount int64) (int64, bool, error) {
	const op = "payment.AddFunds"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to add some funds")

	balance, success, err := p.fundsAdder.AddFunds(ctx, email, amount)
	if err != nil {
		if errors.Is(err, storage.ErrCardNotFound) {
			log.Error("card not found", sl.Err(err))

			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, ErrNotFound)
		}

		log.Error("failed to add some funds", sl.Err(err))

		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("funds added")

	return balance, success, nil
}

func (p *Payment) Pay(ctx context.Context, email string, amount int64) (int64, bool, error) {
	const op = "payment.Pay"

	log := p.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to provide a payment")

	balance, success, err := p.paymentProvider.Pay(ctx, email, amount)
	if err != nil {
		if errors.Is(err, storage.ErrCardNotFound) {
			log.Error("card not found", sl.Err(err))

			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, ErrNotFound)
		}

		if errors.Is(err, storage.ErrNotEnoughFundsToPay) {
			log.Error("not enough funds to pay", sl.Err(err))

			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, ErrNotEnoughFundsToPay)
		}

		log.Error("failed to provide a payment", sl.Err(err))

		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("payment provided")

	return balance, success, nil
}
