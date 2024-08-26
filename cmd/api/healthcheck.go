package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// healthcheckHandler show application information.
func (app *application) healthcheckHandler(ctx *gin.Context) {
	rsp := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		}}

	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
