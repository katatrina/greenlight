package db

import (
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

// MarshalJSON return the JSON-encoded value for the movie runtime,
// in our case, it will return a string in the format "<runtime> mins".
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it in double quotes.
	// It need to be surrounded by double quotes in order to be a valid *JSON string*.
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// Implement a UnmarshalJSON() method on the Runtime type so that is satifies the
// json.Unmarshaler interface. IMPORTANT: because UnmarshalJSON() needs to modify the
// receiver (our Runtime type), we must use a pointer receiver for this to work correctly.
// Otherwise, we will only modiifying a copy (which is then discarded when this method returns).
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// We expect that the incoming JSON value will be a string in the format
	// "<runtime> mins", and the first thing we need to do is remove the surrounding
	// double-quotes from this string. If we can't unquote it, then we return the ErrInvalidRuntimeFormat.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number.
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format.
	// If it isn't, we return the ErrInvalidFormat error again.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Otherwise, parse the string containing the number into an int32. Again, if this
	// fails, we return the ErrInvalidRuntimeFormat error.
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to a Runtime type and assign this to the receiver. Note that we
	// use the * operator to dereference the receiver (which is a pointer to a runtime type)
	// in order to set the underlying value of the pointer.
	*r = Runtime(i)

	return nil
}
