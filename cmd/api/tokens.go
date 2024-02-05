package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/katatrina/greenlight/internal/validator"
	"github.com/pascaldekloe/jwt"
	"net/http"
	"strconv"
	"time"
)

type createAuthenticationTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req createAuthenticationTokenRequest

	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	ValidateEmail(v, req.Email)
	ValidatePasswordPlaintext(v, req.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.store.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	match, err := isPasswordMatched(req.Password, user.PasswordHash)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	var payload jwt.Claims
	payload.Subject = strconv.FormatInt(user.ID, 10)
	payload.Issued = jwt.NewNumericTime(time.Now())
	payload.NotBefore = jwt.NewNumericTime(time.Now())
	payload.Expires = jwt.NewNumericTime(time.Now().Add(12 * time.Hour))
	payload.Set = map[string]interface{}{
		"role": user.Role,
	}

	token, err := payload.HMACSign(jwt.HS256, []byte(app.config.JWTSecretKey))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"authentication_token": string(token)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
