package main

import (
	"database/sql"
	"errors"
	"github.com/katatrina/greenlight/util"
	"github.com/pascaldekloe/jwt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (app *application) authenticate(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			err := errors.New("authorization header is not provided")
			app.errorResponse(w, r, http.StatusUnauthorized, err.Error())
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		payload, err := jwt.HMACCheck([]byte(token), []byte(app.config.JWTSecretKey))
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		if !payload.Valid(time.Now()) {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		userId, err := strconv.ParseInt(payload.Subject, 10, 64)
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.store.GetUserByID(r.Context(), userId)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}

			return
		}

		r = app.contextSetUser(r, &user)

		next.ServeHTTP(w, r)
	}
}

func (app *application) requireManager(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.Role != util.ManagerRole {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}
