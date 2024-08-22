package validator

import (
	"errors"
	"regexp"
)

var (
	isValidUserName = regexp.MustCompile(`^[a-zA-Z0-9\s]+$`).MatchString
	isValidEmail    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString
)

func ValidateUserName(value string) error {
	if err := ValidateStringLength(value, 3, 50); err != nil {
		return err
	}

	if !isValidUserName(value) {
		return errors.New("must contain only letters, numbers, and spaces")
	}

	return nil
}

func ValidateUserEmail(value string) error {
	if err := ValidateStringLength(value, 6, 100); err != nil {
		return err
	}

	if !isValidEmail(value) {
		return errors.New("must be a valid email address")
	}

	return nil
}

func ValidateUserPasswordPlaintext(value string) error {
	if err := ValidateStringLength(value, 8, 72); err != nil {
		return err
	}

	// TODO: Add more password validation rules.

	return nil
}
