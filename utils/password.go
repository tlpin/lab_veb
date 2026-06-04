package utils

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	passwordLetterRegex = regexp.MustCompile(`[A-Za-z]`)
	passwordDigitRegex  = regexp.MustCompile(`[0-9]`)
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if !passwordLetterRegex.MatchString(password) {
		return errors.New("password must contain at least one letter")
	}
	if !passwordDigitRegex.MatchString(password) {
		return errors.New("password must contain at least one digit")
	}
	return nil
}
