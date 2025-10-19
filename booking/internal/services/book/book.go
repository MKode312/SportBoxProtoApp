package book

import (
	"booking/internal/lib/logger/sl"
	"booking/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrAlreadyBooked = errors.New("this box is already booked for this time")
)

type Book struct {
	log    *slog.Logger
	booker Booker
}

func (b *Book) Book(ctx context.Context, email string, boxName string, timeStart string, timeHrs int64, timeMins int64) (reserveID int64, success bool, err error) {
	const op = "book.BookBox"

	log := b.log.With(slog.String("op", op))

	log.Info("booking a box")

	resID, success, err := b.booker.BookABox(ctx, email, boxName, timeStart, timeHrs, timeMins)
	if err != nil {
		log.Error("failed to book a box", sl.Err(err))
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully booked a box")

	return resID, success, nil
}

type TimeCheck struct {
	log         *slog.Logger
	timeChecker TimeChecker
}

type BookCheck struct {
	log             *slog.Logger
	bookingsChecker BookingsChecker
}

type Booker interface {
	BookABox(ctx context.Context, email string, boxName string, startTime string, timeMins int64, timeHrs int64) (resID int64, success bool, err error)
}

type TimeChecker interface {
	TimeCheck(ctx context.Context, timeNow int64) (success bool, err error)
}

type BookingsChecker interface {
	IsNotBooked(ctx context.Context, boxName string, startTime int64, expirationTime int64) (success bool, err error)
}

func NewBooker(log *slog.Logger, booker Booker) *Book {
	return &Book{
		log:    log,
		booker: booker,
	}
}


func (t *TimeCheck) TimeCheck(ctx context.Context, timeNow int64) (bool, error) {
	const op = "book.TimeCheck"

	log := t.log.With(slog.String("op", op))

	log.Info("checking time intervals")

	success, err := t.timeChecker.TimeCheck(ctx, timeNow)
	if err != nil {
		t.log.Error("failed to check time intervals", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("time intervals checked")

	return success, nil
}

func (b *BookCheck) IsNotBooked(ctx context.Context, boxName string, startTime int64, expirationTime int64) (bool, error) {
	const op = "book.IsBooked"

	log := b.log.With(slog.String("op", op))

	log.Info("checking booking")

	isNotBooked, err := b.bookingsChecker.IsNotBooked(ctx, boxName, startTime, expirationTime)
	if err != nil {
		if errors.Is(err, storage.ErrBookingIntervalsCrossed) {
			b.log.Error("booking intervals crossed", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrAlreadyBooked)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("this box is already booked for this time", slog.String("boxNmae", boxName))

	return isNotBooked, nil
}
