package book

import (
	"booking/internal/lib/logger/sl"
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

type Booker interface {
	BookABox(ctx context.Context, email string, boxName string, startTime string, timeMins int64, timeHrs int64) (resID int64, success bool, err error)
}

func NewBooker(log *slog.Logger, booker Booker) *Book {
	return &Book{
		log:    log,
		booker: booker,
	}
}

func (b *Book) Book(ctx context.Context, email string, boxName string, timeStart string, timeHrs int64, timeMins int64) (reserveID int64, success bool, err error) {
	const op = "book.BookBox"

	log := b.log.With(slog.String("op", op))

	log.Info("booking a box")

	resID, success, err := b.booker.BookABox(ctx, email, boxName, timeStart, timeHrs, timeMins)
	if err != nil {
		if err.Error() == ErrAlreadyBooked.Error() {
			log.Error("this box is already booked")
			return 0, false, fmt.Errorf("%s: %w", op, ErrAlreadyBooked)
		}
		log.Error("failed to book a box", sl.Err(err))
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully booked a box")

	return resID, success, nil
}
