package jwtValidation

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyJWTToken(tokenString string) (error) {

	const op = "lib.jwt.validation.VerifyJWTToken"

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("APP_SECRET")), nil
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if token.Valid {
		return nil
	}

	return fmt.Errorf("invalid token")
}