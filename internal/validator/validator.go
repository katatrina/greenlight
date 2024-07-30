package validator

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	isValidMovieTitle = regexp.MustCompile(`^[a-zA-Z0-9\s]+$`).MatchString
)

// Violations is a map of validation errors.
type Violations map[string]string

// New return an empty validation errors map.
func New() Violations {
	return make(map[string]string, 0)
}

// AddError add a new validation error message to the map.
func (v Violations) AddError(field, message string) {
	v[field] = message
}

// Empty check if the validation errors map is empty.
func (v Violations) Empty() bool {
	return len(v) == 0
}

// ValidateStringLength checks if a string value is between a minimum and maximum length.
func ValidateStringLength(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d-%d characters", minLength, maxLength)
	}

	return nil
}

func ValidateMovieTitle(value string) error {
	if err := ValidateStringLength(value, 3, 100); err != nil {
		return err
	}

	if !isValidMovieTitle(value) {
		return errors.New("must contain only letters, numbers, and spaces")
	}

	return nil
}

func ValidateMovieYear(value int32) error {
	now := time.Now()
	if value < 1888 || value > int32(now.Year()) {
		return fmt.Errorf("must be between 1888 and %d", now.Year())
	}

	return nil
}

func ValidateMovieRuntime(value int32) error {
	if value < 1 || value > 300 {
		return errors.New("must be between 1 and 300 minutes")
	}

	return nil
}

func ValidateMovieGenres(genres []string) error {
	if len(genres) < 1 || len(genres) > 5 {
		return errors.New("must contain between 1 and 5 genres")
	}

	firstGenre := genres[0]
	for i := 1; i < len(genres); i++ {
		if genres[i] == firstGenre {
			return errors.New("must not contain duplicate genres")
		}
	}

	return nil
}
