package main

import (
	"context"
	"fmt"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"github.com/katatrina/greenlight/internal/validator"
	"net/http"
	"time"
)

func validateMovie(v *validator.Validator, movie db.Movie) {
	// Use the Check() method to execute our validation checks. This will add the
	// provided key and error message to the "errors" map if the check does not evaluate
	// to true. For example, in the first line here we "check that the title is not
	// equal to the empty string". In the second, we "check that the length of the title
	// is less than or equal to 500 bytes" and so on.

	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	// Note that we're using the Unique helper in the line below to check that all
	// values in the input.Genres slice are unique.
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

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

	movie := db.Movie{
		Title:   req.Title,
		Year:    req.Year,
		Runtime: int32(req.Runtime),
		Genres:  req.Genres,
	}

	// Initialize a new Validator instance for validating data from request body.
	// Note: Validating is the process after decoding successfully.
	// The struct tags cannot validate that a specified field is provided or not,
	// also in the case of providing an empty value.
	// So here, if any field is not provided in the request body, it is empty after decoding to Go struct.
	// And we must validate it manually.
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if
	// any of the checks fail.
	if validateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movie, err = app.store.CreateMovie(context.Background(), db.CreateMovieParams{
		Title:   movie.Title,
		Year:    movie.Year,
		Runtime: movie.Runtime,
		Genres:  movie.Genres,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": newShowMovieResponse(movie)}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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

	// What will we do next after receiving JSON in the response body?
}
