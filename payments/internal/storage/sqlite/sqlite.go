package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"payments/internal/storage"

	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	emptyBalanceValue = -1
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

func (s *Storage) AddCard(ctx context.Context, email string, phoneNumberHash []byte, cardNumberHash []byte, cvcHash []byte) (bool, error) {
	const op = "storage.sqlite.AddCard"

	stmt, err := s.db.Prepare("INSERT INTO cards(email, phone_numberHash, card_numberHash, cvcHash) VALUES(?, ?, ?, ?)")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, email, phoneNumberHash, cardNumberHash, cvcHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return false, fmt.Errorf("%s: %w", op, storage.ErrCardExists)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) AddFunds(ctx context.Context, email string, amount int64) (int64, bool, error) {
	const op = "storage,sqlite.AddFunds"

	stmt, err := s.db.Prepare("SELECT balance FROM cards WHERE email = ?")
	if err != nil {
		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var balance int64
	err = row.Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, storage.ErrCardNotFound)
		}

		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	balance += amount

	stmt2, err := s.db.Prepare("UPDATE cards SET balance = ? WHERE email = ?")
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt2.Close()

	_, err = stmt2.ExecContext(ctx, balance, email)
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	return balance, true, nil
}

func (s *Storage) GetCard(ctx context.Context, email string) (string, string, error) {
	const op = "storage.sqlite.GetCard"

	stmt, err := s.db.Prepare("SELECT card_numberHash, phone_numberHash FROM cards WHERE email = ?")
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var cardNumberHash []byte
	var phoneNumberHash []byte
	err = stmt.QueryRowContext(ctx, email).Scan(&cardNumberHash, &phoneNumberHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("%s: %w", op, storage.ErrCardNotFound)
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return string(cardNumberHash), string(phoneNumberHash), nil
}

func (s *Storage) Pay(ctx context.Context, email string, amount int64) (int64, bool, error) {
	const op = "storage.sqlite.Pay"

	stmt, err := s.db.Prepare("SELECT balance FROM cards WHERE email = ?")
	if err != nil {
		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var balance int64
	err = stmt.QueryRowContext(ctx, email).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, storage.ErrCardNotFound)
		}
		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
	}

	if balance >= amount {
		balance -= amount

		stmt, err = s.db.Prepare("UPDATE cards SET balance = ? WHERE email = ?")
		if err != nil {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx, balance, email)
		if err != nil {
			return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, err)
		}

		return balance, true, nil
	} else {
		return emptyBalanceValue, false, fmt.Errorf("%s: %w", op, storage.ErrNotEnoughFundsToPay)
	}
}
