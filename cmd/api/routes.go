package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Convert the notFoundResponse helper to a http.Handler using the
	// http.HandlerFunc adapter, and then set it as the custom error handler for 404
	// Not Found response.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Likewise, convert the methodNotAllowedResponse helper to a http.Handler and set
	// it as the custom error handler for 405 Method Not Allowed response.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the relevant methods, URL patterns, and handler functions for our
	// endpoints using the HandlerFunc() method.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// Return the httprouter instance.
	return router
}
