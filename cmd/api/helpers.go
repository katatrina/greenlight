package main

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// readIDParam retrieve the "id" URL parameter from the provided request context,
// then convert it into an integer and return it.
// If the parameter couldn't be converted, or is less than 1, return 0 and an error.
func (app *application) readIDParam(ctx *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}
