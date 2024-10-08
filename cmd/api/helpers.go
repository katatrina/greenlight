package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type envelope map[string]any

// readIDParam retrieve the "id" URL parameter from the provided request context,
// then convert it into an integer and return it.
// If the parameter couldn't be converted, or is less than 1, return 0 and an error.
func (app *application) readIDParam(ctx *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// writeJSON take the request context, the HTTP status code to send, the data must be a struct or a map to be encoded to JSON response body, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(ctx *gin.Context, statusCode int, data any, headers map[string]string) {
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(statusCode, data)
}

// readJSON decode the request body into destination struct.
// It asserts the body contains a valid JSON object, and returns an error if not.
func (app *application) readJSON(ctx *gin.Context, destinaton any) error {
	// Limit the size of our request body to 1MB.
	maxBytes := 1_048_576
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, int64(maxBytes))

	// TODO: Disallow unknown fields in body (currently not supported by GIN)
	/* TODO: Differentiate between missing, null, and empty values
	1. Use the binding:"required" struct tag for each field
		=> error is returned by ShouldBindJSON
		=> cannot custom error string (urly error string)
		=> cannot catch multiple errors at once (multiple not-provided errors)
		 => The clients have to sent another request to see remaining errors if have.
		=> but can reduce boilerplate code
	2. Use pointers for each data types. For example: string => *string, int64 => *int64, ...
		=> introduce lots of boilerplate code
		=> not idiomatic
		=> but we completely controll over the validation process.
	3. ...
	*/

	err := ctx.ShouldBindJSON(destinaton)
	if err != nil {
		var (
			syntaxError        *json.SyntaxError        // There is a syntax problem with the JSON being decoded.
			unmarshalTypeError *json.UnmarshalTypeError // A JSON value is not appropriate for the destination Go data type.
			/*
				The destination is not valid (it must be a non-nil pointer).
				This is actually a problem with our application code, not the JSON itself.
			*/
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError // The request body exceeded our size limit.
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// TODO: Check if the request body only contains a single json value

	return nil
}

// readQueryParams decode the query string parameters into destination struct.
//
// Any mismatch-related data type errors will be catched here.
func (app *application) readQueryParams(ctx *gin.Context, destination any) error {
	err := ctx.ShouldBindQuery(destination)
	if err != nil {
		// TODO: Handle error more gracefully
		return err
	}

	return nil
}

// background runs the fn function in another goroutine.
func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if panicVal := recover(); panicVal != nil {
				app.logger.Error(fmt.Sprintf("%v", panicVal))
			}
		}()

		fn()
	}()
}
