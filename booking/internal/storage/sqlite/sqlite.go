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

	stmt, err := s.db.Prepare("INSERT INTO bookings(email, boxName, startsAt, expiresAt) VALUES(?, ?, ?, ?)")
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	dur := time.Duration(timeHrs)*time.Hour + time.Duration(timeMins)*time.Minute

	startsAt, err := time.Parse(time.UnixDate, startTime)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	timeEnd := startsAt.Add(dur).Local().Unix()

	res, err := stmt.ExecContext(ctx, email, boxName, startTime, timeEnd)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	return id, true, nil
}

func (s *Storage) TimeCheck(ctx context.Context, timeNow int64) (bool, error) {
	const op = "storage.sqlite.TimeCheck"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM bookings WHERE expiresAt < ?
    `, timeNow)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) IsNotBooked(ctx context.Context, boxName string, startTime int64, expirationTime int64) (bool, error) {
	const op = "storage.sqlite.IsBooked"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	row := tx.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM bookings WHERE boxName = ? AND (startsAt < ? AND expiresAt > ?) 
    `, boxName, expirationTime, startTime)
	var count int
	if err := row.Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return false, fmt.Errorf("%s: %w", op, storage.ErrBookingNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if count > 0 {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return false, fmt.Errorf("%s: %w", op, err)
		}
		return true, fmt.Errorf("%s: %w", op, storage.ErrBookingIntervalsCrossed)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}
