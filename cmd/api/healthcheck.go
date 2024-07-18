package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// healthcheckHandler show application information.
func (app *application) healthcheckHandler(ctx *gin.Context) {
	ctx.Writer.Write([]byte("status: available\n"))
	ctx.Writer.Write([]byte(fmt.Sprintf("environment: %s\n", app.config.env)))
	ctx.Writer.Write([]byte(fmt.Sprintf("version: %s\n", version)))
}
