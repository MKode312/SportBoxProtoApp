package sqlite

import (
	"booking/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) BookABox(ctx context.Context, email string, boxName string, startTime string, timeHrs int64, timeMins int64) (resID int64, success bool, err error) {
	const op = "storage.sqlite.BookABox"

	dur := time.Duration(timeHrs)*time.Hour + time.Duration(timeMins)*time.Minute

	startsAt, err := time.Parse(time.UnixDate, startTime)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	timeEnd := startsAt.Add(dur).Local().Unix()

	isNotBooked, err := s.IsNotBooked(ctx, boxName, startsAt.Unix(), timeEnd)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("isNotBooked:", isNotBooked)

	if isNotBooked {
		fmt.Println("Попытка вставить запись")
		stmt, err := s.db.Prepare("INSERT INTO bookings(email, boxName, startsAt, expiresAt) VALUES(?, ?, ?, ?)")
		if err != nil {
			fmt.Println("Ошибка подготовки запроса:", err)
			return 0, false, fmt.Errorf("%s: %w", op, err)
		}

		res, err := stmt.ExecContext(ctx, email, boxName, startsAt.Unix(), timeEnd)
		if err != nil {
			fmt.Println("Ошибка выполнения вставки:", err)
			return 0, false, fmt.Errorf("%s: %w", op, err)
		} else {
			fmt.Println("Вставка выполнена успешно")
		}

		id, err := res.LastInsertId()
		if err != nil {
			fmt.Println("Ошибка получения last insert id:", err)
			return 0, false, fmt.Errorf("%s: %w", op, err)
		}
		fmt.Printf("Последний вставленный ID: %d\n", id)
		return 1000 * id, true, nil
	} else {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}
}

func (s *Storage) TimeCheck(ctx context.Context, timeNow int64) (bool, error) {
	const op = "storage.sqlite.TimeCheck"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM bookings WHERE expiresAt < ?
    `, timeNow)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) StartDbChecker(ctx context.Context, interval int64) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		ticker := time.NewTicker(time.Second * time.Duration(interval))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().Unix()
				if _, err := s.TimeCheck(ctx, now); err != nil {
					errCh <- err
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return errCh
}

func (s *Storage) IsNotBooked(ctx context.Context, boxName string, startTime int64, expirationTime int64) (bool, error) {
	const op = "storage.sqlite.IsBooked"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	row := tx.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM bookings WHERE boxName = ? AND (startsAt <= ? AND expiresAt >= ?) 
    `, boxName, expirationTime, startTime)
	var count int
	fmt.Printf("Проверка брони для boxName=%s, startTime=%d, expirationTime=%d\n", boxName, startTime, expirationTime)
	if err := row.Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrBookingNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
		fmt.Printf("Результат COUNT=%d\n", count)
	if count > 0 {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		return false, fmt.Errorf("%s: %w", op, storage.ErrAlreadyBooked)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) CancelBooking(ctx context.Context, email string, bookingID int64) (refundedAmount int64, success bool, err error) {
	const op = "storage.sqlite.CancelBooking"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	// First, verify the booking exists and belongs to the user
	row := tx.QueryRowContext(ctx, `
		SELECT email, boxName, startsAt, expiresAt FROM bookings WHERE bookingId = ?
	`, bookingID/1000) // Convert from external ID to internal ID
	
	var (
		bookingEmail string
		boxName      string
		startsAt     int64
		expiresAt    int64
	)

	if err := row.Scan(&bookingEmail, &boxName, &startsAt, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, fmt.Errorf("%s: %w", op, storage.ErrBookingNotFound)
		}
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	if bookingEmail != email {
		return 0, false, fmt.Errorf("%s: %w", op, storage.ErrNotYourBooking)
	}

	// Calculate refund amount based on remaining time
	now := time.Now().Unix()
	if now >= expiresAt {
		return 0, false, fmt.Errorf("%s: booking has already expired", op)
	}

	remainingTime := expiresAt - now
	totalTime := expiresAt - startsAt
	refundedAmount = (remainingTime * 100) / totalTime // Calculate percentage of time remaining

	// Delete the booking
	result, err := tx.ExecContext(ctx, "DELETE FROM bookings WHERE bookingId = ?", bookingID/1000)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return 0, false, fmt.Errorf("%s: %w", op, storage.ErrBookingNotFound)
	}

	if err := tx.Commit(); err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	return refundedAmount, true, nil
}