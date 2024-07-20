package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type envelop map[string]any

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

// writeJSON take the request context, the HTTP status code to send, the data must be a struct or a map to be encoded to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(ctx *gin.Context, statusCode int, data any, headers map[string]string) {
	for key, value := range headers {
		ctx.Header(key, value)
	}

	ctx.JSON(statusCode, data)
}

// readJSON decode the request body into destination struct.
func (app *application) readJSON(ctx *gin.Context, destinaton any) error {
	err := ctx.ShouldBindJSON(destinaton)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

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

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	return nil
}
