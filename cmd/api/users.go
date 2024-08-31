package main

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/mailer"
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

	// Parse request body
	err := app.readJSON(ctx, &req)
	if err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate request body
	violations := validateRegisterUserRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// NOTE: If the email is already exists, the calculation of hashing password is unnecessary and can make the response slower.
	// But querying the database to check if the email already exists before hashing the password is also not a good idea.

	// Generate a hashed password from the plaintext password.
	hashedPassword, err := util.HashPassword(*req.Password)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	arg := db.RegisterUserTxParams{
		Name:           *req.Name,
		Email:          *req.Email,
		HashedPassword: hashedPassword,
		Permissions:    []string{movieReadPermissionCode},
	}

	// Try to create a new user account with the default permissions.
	user, err := app.store.RegisterUserTx(ctx, arg)
	if err != nil {
		// If the email address is already in use, return a 409 status code and a JSON response.
		if db.ErrorCode(err) == db.UniqueViolation && db.IsContainErrorMessage(err, "users_email_key") {
			app.integrityConstraintViolationResponse(ctx, "a user with this email address already exists")
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// After the user record has been created, generate a new activation token for the user.
	tokenPlaintext, _, err := app.store.GenerateToken(ctx, db.GenerateTokenParams{
		UserID:   user.ID,
		Duration: 3 * 24 * time.Hour,
		Scope:    db.ScopeActivation,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// Send the user a welcome email with the activation token.
	app.background(func() {
		header := mailer.EmailHeader{
			Subject: "Welcome to Greenlight!",
			To:      []string{user.Email},
		}

		data := map[string]any{
			"userID":          user.ID,
			"activationToken": tokenPlaintext,
		}

		err := app.mailer.SendEmail(header, data, "user_welcome.html")
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	rsp := envelope{"user": user}
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

	// Parse request body
	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate request body
	violations := validateActivateUserRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Generate a SHA-256 hash of the plaintext token string.
	tokenHash := sha256.Sum256([]byte(*req.TokenPlainText))

	// Retrieve the details of the user associated with the token hash.
	// If no matching row is found, then we let the client know that the token they provided is not valid.
	user, err := app.store.GetUserByToken(ctx, db.GetUserByTokenParams{
		Hash:  tokenHash[:],
		Scope: db.ScopeActivation,
	})
	if err != nil {
		// If no matching row is found, then we let the client know that the token they provided is not valid.
		if errors.Is(err, db.ErrRecordNotFound) {
			violations.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(ctx, violations)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// Return an error if the user account has already been activated.
	if user.Activated {
		violations.AddError("email", "user has already been activated")
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Try to activate the user account.
	activatedUser, err := app.store.ActivateUserTx(ctx, db.ActivateUserParams{
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

	// Send the activated user details to the client in a JSON response.
	rsp := envelope{"user": activatedUser}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}

type resetUserPasswordRequest struct {
	TokenPlaintext *string `json:"token"`
	NewPassword    *string `json:"password"`
}

func validateResetUserPasswordRequest(req *resetUserPasswordRequest) validator.Violations {
	violations := validator.New()

	if req.TokenPlaintext == nil {
		violations.AddError("token", "must be provided")
	} else if err := validator.ValidateTokenPlaintext(*req.TokenPlaintext); err != nil {
		violations.AddError("token", err.Error())
	}

	if req.NewPassword == nil {
		violations.AddError("password", "must be provided")
	} else if err := validator.ValidateUserPasswordPlaintext(*req.NewPassword); err != nil {
		violations.AddError("password", err.Error())
	}

	return violations
}

// resetUserPasswordHandler update the user's password.
func (app *application) resetUserPasswordHandler(ctx *gin.Context) {
	var req resetUserPasswordRequest

	// Parse request body
	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate request body
	violations := validateResetUserPasswordRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Generate a SHA-256 hash of the plaintext token string.
	tokenHash := sha256.Sum256([]byte(*req.TokenPlaintext))

	// Retrieve the details of the user associated with the hashed token.
	user, err := app.store.GetUserByToken(ctx, db.GetUserByTokenParams{
		Hash:  tokenHash[:],
		Scope: db.ScopePasswordReset,
	})
	if err != nil {
		// If no matching row is found, then we let the client know that the token they provided is not valid.
		if errors.Is(err, db.ErrRecordNotFound) {
			violations.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(ctx, violations)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// Generate a new hashed password from the input password.
	hashedPassword, err := util.HashPassword(*req.NewPassword)
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// Set the new hashed password for the user.
	err = app.store.ResetUserPasswordTx(ctx, db.ResetUserPasswordTxParams{
		UserID:         user.ID,
		HashedPassword: hashedPassword,
		Version:        user.Version,
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

	// Send the client a 200 OK response with a success message.
	rsp := envelope{"message": "your password was successfully reset"}
	app.writeJSON(ctx, http.StatusOK, rsp, nil)
}
