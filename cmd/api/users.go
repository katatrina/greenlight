package main

import (
	"context"
	"errors"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"github.com/katatrina/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type registerUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerUserResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"created_at"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest

	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if ValidateUser(v, req); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	hashedPassword, err := generatePasswordHash(req.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	arg := db.CreateUserParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Activated:    false,
	}

	user, err := app.store.CreateUser(context.Background(), arg)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "users_email_key"`):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	resp := registerUserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Activated: user.Activated,
		CreatedAt: user.CreatedAt,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": resp}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func generatePasswordHash(plaintextPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
}

// IsPasswordMatched checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func isPasswordMatched(plaintextPassword string, hashedPassword []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user registerUserRequest) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	ValidatePasswordPlaintext(v, user.Password)
}
