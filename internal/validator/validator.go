package validator

import (
	"fmt"
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
