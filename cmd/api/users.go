package main

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/util"
	"github.com/katatrina/greenlight/internal/validator"
)

type registerUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

func validateRegisterUserRequest(req *registerUserRequest) validator.Violations {
	violations := validator.New()

	if req.Name == nil {
		violations.AddError("name", "must be provided")
	} else if err := validator.ValidateUserName(*req.Name); err != nil {
		violations.AddError("name", err.Error())
	}

	if req.Email == nil {
		violations.AddError("email", "must be provided")
	} else if err := validator.ValidateUserEmail(*req.Email); err != nil {
		violations.AddError("email", err.Error())
	}

	if req.Password == nil {
		violations.AddError("password", "must be provided")
	} else if err := validator.ValidateUserPasswordPlaintext(*req.Password); err != nil {
		violations.AddError("password", err.Error())
	}

	return violations
}

// registerUserHandler create a new user account.
func (app *application) registerUserHandler(ctx *gin.Context) {
	var req registerUserRequest

	err := app.readJSON(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	violations := validateRegisterUserRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// TODO: If the email is already exists, the calculation of hashing password is unnecessary and can make the response slower.

	hashedPassword, err := util.HashPassword(*req.Password)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	arg := db.CreateUserParams{
		Name:           *req.Name,
		Email:          *req.Email,
		HashedPassword: hashedPassword,
		Activated:      false,
	}

	user, err := app.store.CreateUser(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation && db.IsContainErrorMessage(err, "users_email_key") {
			app.integrityConstraintViolationResponse(ctx, "a user with this email address already exists")
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// After the user record has been created in the database, generate a new activation token for the user.
	tokenPlaintext, _, err := app.store.GenerateToken(ctx, user.ID, 3*24*time.Hour, db.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": tokenPlaintext,
			"userID":          user.ID,
		}

		err = app.mailer.SendEmail("Welcome <3!!!", data, []string{user.Email}, nil, nil, nil, "user_welcome.html")
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	rsp := envelop{"user": user}
	// 202 Accepted status code indicates that the request has been accepted for processing, but
	// the processing has not been completed.
	app.writeJSON(ctx, http.StatusAccepted, rsp, nil)
}

type activateUserRequest struct {
	TokenPlainText *string `json:"token"`
}

func validateActivateUserRequest(req *activateUserRequest) validator.Violations {
	violations := validator.New()

	if req.TokenPlainText == nil {
		violations.AddError("token", "must be provided")
	} else if err := validator.ValidateTokenPlaintext(*req.TokenPlainText); err != nil {
		violations.AddError("token", err.Error())
	}

	return violations
}

// activateUserHandler update the user's activated field to true.
func (app *application) activateUserHandler(ctx *gin.Context) {
	var req activateUserRequest

	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	violations := validateActivateUserRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	tokenHash := sha256.Sum256([]byte(*req.TokenPlainText))

	// Retrieve the details of the user associated with the token.
	// If no matching row is found, then we let the client know that the token they provided is not valid.
	user, err := app.store.GetUserByToken(ctx, db.GetUserByTokenParams{
		Hash:  tokenHash[:],
		Scope: db.ScopeActivation,
	})
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			violations.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(ctx, violations)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	activatedUser, err := app.store.ActivateUser(ctx, db.ActivateUserParams{
		UserID:  user.ID,
		Version: user.Version,
	})
	if err != nil {
		// If no matching row could be found, we know the user's version has changed
		// (or the record has been deleted) and we invoke the editConflictResponse method.
		if errors.Is(err, db.ErrRecordNotFound) {
			app.editConflictResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// If everything went successfully, then we delete all activation tokens of the user.
	err = app.store.DeleteUserTokens(ctx, db.DeleteUserTokensParams{
		UserID: activatedUser.ID,
		Scope:  db.ScopeActivation,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// TODO: Maybe using an atomic transaction for all above queries could be better.

	// Send the activated user details to the client in a JSON response.
	rsp := envelop{"user": activatedUser}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
