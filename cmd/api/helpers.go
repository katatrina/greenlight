package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelope map[string]any

// writeJSON is a helper that writes the given data to the client in JSON format with the appropriate headers and status code.
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	jsonString, err := json.Marshal(data)
	if err != nil {
		return err
	}

	jsonString = append(jsonString, '\n')

	// Go doesn't throw an error if you try to range over a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Set the "Content-Type: application/json" header on the response. If you forget to
	// this, Go will default to sending a "Content-Type: text/plain; charset=utf-8"
	// header instead.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonString)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder; and call the method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field that cannot be mapped to the target destination, the decoder will return
	// an error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination.
	// Note: When decoding, Golang will check the general form of JSON,
	// then it continues to decode each field in JSON request body.
	// So there are errors that belong to the general form of JSON
	// such as io.EOF, io.ErrUnexpectedEOF, json.SyntaxError, http.MaxBytesError.
	// And there are errors that belong to the specific field in JSON request body
	// such as json.UnmarshalTypeError, json.InvalidUnmarshalError, etc.
	// The errors that belong to the general form of JSON will be checked first.
	// If there is no error, Golang will continue to check the errors that belong to the specific field.
	// If there is an error, Golang will stop decoding and return that error.
	err := dec.Decode(dst)
	if err != nil {
		// If there is an error during encoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// Use the errors.As() function to check whether the error has the type *json.SyntaxError.
		// If it does, then return a plain-english error message which includes the location of the problem
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON.
		// So we check for this using errors.Is() and return a generic error message.
		// There is an open issue regarding this at https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination. If the error relates
		// to a specific field, then we include that in our error message to make it
		// easier for the client to debug.
		// Note: this error will never be caught if we customized the Unmarshal on runtime field.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty.
		// We check for this with errors.Is() and return a plain-english error message instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format 'json: unknown field "field"'.
		// We check for this, extract the field name from the error, and interpolate it into our custom error message.
		// Note that there's an open issue at https://github.com/golang/go/issues/29035 regarding turning this
		// into a distinct error type in the future.
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// Use the error.As() function to check whether the error has the type *http.MaxByteError.
		// If it does, then it means the request body exceeded our size limit of 1MB,
		// and we return a clear error message.
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// A json.InvalidUnmarshalError error will be returned if we pass something
		// that is not a non-nil pointer to Decode(). We catch this and panic,
		// rather than returning an error to our handler. At the end of this chapter,
		// we'll talk about panicking versus returning errors, and discuss why it's an
		// appropriate thing to do in this specific situation.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct as the destination.
	// If the request body only contained a single JSON value, this will return an io.EOF error.
	// So if we get anything else, we know that there is additional data in the request body,
	// and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contains a single JSON value")
	}

	return nil
}
