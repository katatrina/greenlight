package main

import (
	"fmt"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"net/http"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

type showMovieResponse struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Year    int32    `json:"year,omitempty"`
	Runtime int32    `json:"runtime,omitempty"`
	Genres  []string `json:"genres,omitempty"`
	Version int32    `json:"version"`
}

func newShowMovieResponse(movie db.Movie) showMovieResponse {
	return showMovieResponse{
		ID:      movie.ID,
		Title:   movie.Title,
		Year:    movie.Year,
		Runtime: movie.Runtime,
		Genres:  movie.Genres,
		Version: movie.Version,
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	movie := db.Movie{
		ID:      id,
		Title:   "Casablanca",
		Runtime: 102,
		Genres:  []string{"drama", "romance", "war"},
		Version: 1,
	}

	rsp := newShowMovieResponse(movie)

	err = app.writeJSON(w, http.StatusOK, rsp, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
