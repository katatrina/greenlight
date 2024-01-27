package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"github.com/katatrina/greenlight/internal/validator"
	"net/http"
	"time"
)

func validateMovie(v *validator.Validator, movie *db.Movie) {
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

	// Checkpoint: Decoding successfully

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
	/*
		Nah bro, Gin - a web framework can validate by struct tags.
	*/
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if
	// any of the checks fail.
	if validateMovie(v, &movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Checkpoint: Validating user input successfully

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

	movie, err := app.store.GetMovie(context.Background(), id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	resp := envelope{"movie": newShowMovieResponse(movie)}

	err = app.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// What will we do next after receiving JSON in the response body?
}

type updateMovieRequest struct {
	Title   *string     `json:"title"` // This will be a nil if there is no corresponding key in the JSON.
	Year    *int32      `json:"year"`
	Runtime *db.Runtime `json:"runtime"`
	Genres  []string    `json:"genres"` // We don't need to change this because slices already have the zero-value nil.
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.store.GetMovie(context.Background(), id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	var req updateMovieRequest
	err = app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// If the input.Title value is nil then we know that no corresponding "title" key/
	// value pair was provided in the JSON request body. So we move on and leave the
	// movie record unchanged. Otherwise, we update the movie record with the new title
	// value. Importantly, because input.Title is a now a pointer to a string, we need
	// to dereference the pointer using the * operator to get the underlying value
	// before assigning it to our movie record.
	if req.Title != nil {
		movie.Title = *req.Title
	}

	// We also do the same for the other fields in the input struct.
	if req.Year != nil {
		movie.Year = *req.Year
	}

	if req.Runtime != nil {
		movie.Runtime = int32(*req.Runtime)
	}

	if req.Genres != nil {
		movie.Genres = req.Genres
	}

	v := validator.New()
	if validateMovie(v, &movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movie, err = app.store.UpdateMovie(context.Background(), db.UpdateMovieParams{
		Title:   movie.Title,
		Year:    movie.Year,
		Runtime: movie.Runtime,
		Genres:  movie.Genres,
		ID:      movie.ID,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	resp := envelope{"movie": newShowMovieResponse(movie)}

	err = app.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	rowsAffected, err := app.store.DeleteMovie(context.Background(), id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// If no rows were affected, we know that the "movies" table didn't contain a record
	// with the provided ID at the moment we tried to delete it.
	// In that case, we return a not found error.
	if rowsAffected == 0 {
		app.notFoundResponse(w, r)
		return
	}

	resp := envelope{"message": "movie successfully deleted!"}

	err = app.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
