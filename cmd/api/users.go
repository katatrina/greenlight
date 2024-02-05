package main

import (
	"context"
	"errors"
	db "github.com/katatrina/greenlight/internal/db/sqlc"
	"github.com/katatrina/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
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

func (app *application) listUserHandler(w http.ResponseWriter, r *http.Request) {
	g := new(errgroup.Group)

	resp := envelope{"users": []db.User{}, "total_users": 0}

	g.Go(func() error {
		totalUsers, err := app.store.ListTotalUsers(context.Background())
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return err
		}

		resp["total_users"] = totalUsers
		return err
	})

	g.Go(func() error {
		pageID, err := strconv.Atoi(r.URL.Query().Get("page_id"))
		if err != nil || pageID < 1 {
			err := errors.New("page_id must be greater than zero")
			app.notFoundResponse(w, r)
			return err
		}

		pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
		if err != nil || pageSize < 1 || pageSize > 10 {
			err := errors.New("page_size must be between 1 and 10")
			app.notFoundResponse(w, r)
			return err
		}

		arg := db.ListUsersParams{
			Limit:  int32(pageSize),
			Offset: int32((pageID - 1) * pageSize),
		}

		users, err := app.store.ListUsers(context.Background(), arg)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return err
		}
		resp["users"] = users
		return err
	})

	if err := g.Wait(); err == nil {
		err = app.writeJSON(w, http.StatusOK, resp, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
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
