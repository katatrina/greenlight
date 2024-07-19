package data

import (
	"fmt"
	"strconv"
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
