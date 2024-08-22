package util

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the plaintext password using the bcrypt algorithm.
func HashPassword(plaintextPassword string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return []byte(""), err
	}

	return hashedPassword, nil
}

// CheckPassword compares the user's hashed password vs the provided plaintext password.
func CheckPassword(hashedPassword, plaintextPassword []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, plaintextPassword)
	if err != nil {
		return err
	}

	return nil
}
