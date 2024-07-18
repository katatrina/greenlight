package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// createMovieHandler create a new movie.
func (app *application) createMovieHandler(ctx *gin.Context) {
	ctx.Writer.Write([]byte("create a new movie"))
}

// showMovieHandler show the details of a specific movie.
func (app *application) showMovieHandler(ctx *gin.Context) {
	// Try to convert the id string to a base 10 integer (with a bit size of 64)
	movieID, err := app.readIDParam(ctx)
	if err != nil {
		ctx.String(http.StatusNotFound, "404 page not found")
		return
	}

	ctx.Writer.Write([]byte(fmt.Sprintf("show the details of movie %d\n", movieID)))
}
