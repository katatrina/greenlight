package validator

import "errors"

func ValidateTokenPlaintext(token string) error {
	if len(token) != 26 {
		return errors.New("must contain exactly 26 characters")
	}

	// TODO: Add more token validation rules.

	return nil
}