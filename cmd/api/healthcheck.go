package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// healthcheckHandler show application information.
func (app *application) healthcheckHandler(ctx *gin.Context) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	app.writeJSON(ctx, http.StatusOK, data, nil)
}
