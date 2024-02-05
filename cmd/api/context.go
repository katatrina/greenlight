package main

import (
	"context"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *db.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *db.User {
	user, ok := r.Context().Value(userContextKey).(*db.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
