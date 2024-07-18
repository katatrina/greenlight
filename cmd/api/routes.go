package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	// Initialize a new gin router instance.
	router := gin.Default()
	router.HandleMethodNotAllowed = true

	router.GET("/v1/healthcheck", app.healthcheckHandler)
	router.GET("/v1/movies/:id", app.showMovieHandler)
	router.POST("/v1/movies", app.createMovieHandler)

	return router
}
