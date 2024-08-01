package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	// Initialize a new gin router instance.
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.NoMethod(app.methodNotAllowedResponse)
	router.NoRoute(app.notFoundResponse)

	router.GET("/v1/healthcheck", app.healthcheckHandler)

	movieRoutes := router.Group("/v1/movies")
	{
		movieRoutes.POST("", app.createMovieHandler)
		movieRoutes.GET("/:id", app.showMovieHandler)
		movieRoutes.GET("", app.listMoviesHandler)
		movieRoutes.PATCH("/:id", app.updateMovieHandler)
		movieRoutes.DELETE("/:id", app.deleteMovieHandler)
	}

	return router
}
