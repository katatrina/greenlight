package main

import (
	"fmt"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"net/http"
)

type createMovieRequest struct {
	Title   string     `json:"title"`
	Year    int32      `json:"year"`
	Runtime db.Runtime `json:"runtime"`
	Genres  []string   `json:"genres"`
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var req createMovieRequest

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error, we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", req)
}

type showMovieResponse struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Year  int32  `json:"year,omitempty"`
	// Use the Runtime type instead of int32. Note that the omitempty directive will
	// still work on this: if the Runtime field has the underlying value 0, then it will
	// be considered empty and omitted -- and the MarshalJSON() method we just made
	// won't be called at all.
	Runtime db.Runtime `json:"runtime,omitempty"`
	Genres  []string   `json:"genres,omitempty"`
	Version int32      `json:"version"`
}

func newShowMovieResponse(movie db.Movie) *showMovieResponse {
	return &showMovieResponse{
		ID:      movie.ID,
		Title:   movie.Title,
		Year:    movie.Year,
		Runtime: db.Runtime(movie.Runtime),
		Genres:  movie.Genres,
		Version: movie.Version,
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
	}

	movie := db.Movie{
		ID:      id,
		Title:   "Casablanca",
		Year:    0, // this field will be omitted in the response body
		Runtime: 102,
		Genres:  []string{"drama", "romance", "war"},
		Version: 1,
	}

	resp := envelope{"movie": newShowMovieResponse(movie)}

	err = app.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
