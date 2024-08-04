package util

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes the plaintext password using the bcrypt algorithm.
func HashPassword(plaintextPassword string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return []byte(""), err
	}

	return hashedPassword, nil
}
