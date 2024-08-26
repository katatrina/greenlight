package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	movieReadPermissionCode  = "movies:read"
	movieWritePermissionCode = "movies:write"
)

func (app *application) routes() http.Handler {
	// Initialize a new gin router instance.
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	router.NoMethod(app.methodNotAllowedResponse)
	router.NoRoute(app.notFoundResponse)

	router.Use(app.authenticate()) // we want to authenticate user on all requests.

	router.GET("/v1/healthcheck", app.healthcheckHandler)

	movieRoutes := router.Group("/v1/movies", app.requireAuthenticatedUser(), app.requireActivatedUser())
	{
		movieRoutes.POST("", app.requirePermission(movieWritePermissionCode), app.createMovieHandler)
		movieRoutes.GET("/:id", app.requirePermission(movieReadPermissionCode), app.showMovieHandler)
		movieRoutes.GET("", app.requirePermission(movieReadPermissionCode), app.listMoviesHandler)
		movieRoutes.PATCH("/:id", app.requirePermission(movieWritePermissionCode), app.updateMovieHandler)
		movieRoutes.DELETE("/:id", app.requirePermission(movieWritePermissionCode), app.deleteMovieHandler)
	}

	userRoutes := router.Group("/v1/users")
	{
		userRoutes.POST("", app.registerUserHandler)
		userRoutes.PUT("/activated", app.activateUserHandler)
		userRoutes.PUT("/password", app.updateUserPasswordHandler)
	}

	tokenRoutes := router.Group("/v1/tokens")
	{
		tokenRoutes.POST("/authentication", app.createAuthenticationTokenHandler) // login

		tokenRoutes.POST("/password-reset", app.createPasswordResetTokenHandler)
	}

	return router
}
