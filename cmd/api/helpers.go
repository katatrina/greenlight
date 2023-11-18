package main

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
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
