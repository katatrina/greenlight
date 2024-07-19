package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/data"
)

// createMovieHandler create a new movie.
func (app *application) createMovieHandler(ctx *gin.Context) {
	ctx.Writer.Write([]byte("create a new movie"))
}

// showMovieHandler show the details of a specific movie.
func (app *application) showMovieHandler(ctx *gin.Context) {
	// Try to convert the id string to a base 10 integer (with a bit size of 64).
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		app.notFoundResponse(ctx)
		return
	}

	movie := data.Movie{
		ID:        movieID,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	rsp := envelop{"movie": movie}

	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
