package main

import (
	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
)

const (
	userContextKey = "user"
)

func (app *application) contextSetUser(ctx *gin.Context, user *db.User) {
	ctx.Set(userContextKey, user)
}

func (app *application) contextGetUser(ctx *gin.Context) *db.User {
	user, ok := ctx.MustGet(userContextKey).(*db.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
