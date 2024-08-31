package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/mailer"
	"github.com/katatrina/greenlight/internal/util"
	"github.com/katatrina/greenlight/internal/validator"
)

type createAuthenticationTokenRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type createAuthenticationTokenResponse struct {
	TokenPlaintext string    `json:"token"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func validateCreateAuthenticationTokenRequest(req *createAuthenticationTokenRequest) validator.Violations {
	violations := validator.New()

	// validate email
	if req.Email == nil {
		violations.AddError("email", "must be provided")
	} else if err := validator.ValidateUserEmail(*req.Email); err != nil {
		violations.AddError("email", err.Error())
	}

	// validate password
	if req.Password == nil {
		violations.AddError("password", "must be provided")
	} else if err := validator.ValidateUserPasswordPlaintext(*req.Password); err != nil {
		violations.AddError("password", err.Error())
	}

	return violations
}

// createAuthenticationTokenHandler create a new stateful authentication token for the user.
func (app *application) createAuthenticationTokenHandler(ctx *gin.Context) {
	var req createAuthenticationTokenRequest

	// Parse request body
	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate request body
	violations := validateCreateAuthenticationTokenRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Lookup the user record based on the email address. If no matching user was
	// found, we send a 401 Unauthorized to the client.
	user, err := app.store.GetUserByEmail(ctx, *req.Email)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			app.invalidCredentialsResponse(ctx)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	err = util.CheckPassword(user.HashedPassword, []byte(*req.Password))
	if err != nil {
		app.invalidCredentialsResponse(ctx)
		return
	}

	// If the password is correct, we generate a new stateful authentication token
	// with a 24-hour expiry time with the scope 'authentication'.
	tokenPlaintext, token, err := app.store.GenerateToken(ctx, db.GenerateTokenParams{
		UserID:   user.ID,
		Duration: 24 * time.Hour,
		Scope:    db.ScopeAuthentication,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelope{
		"authentication_token": createAuthenticationTokenResponse{
			TokenPlaintext: tokenPlaintext,
			ExpiresAt:      token.ExpiresAt,
		},
	}
	app.writeJSON(ctx, http.StatusCreated, rsp, nil)
}

type createPasswordResetTokenRequest struct {
	Email *string `json:"email"`
}

func validateCreatePasswordResetTokenRequest(req *createPasswordResetTokenRequest) validator.Violations {
	violations := validator.New()

	if req.Email == nil {
		violations.AddError("email", "must be provided")
	} else if err := validator.ValidateUserEmail(*req.Email); err != nil {
		violations.AddError("email", err.Error())
	}

	return violations
}

// createPasswordResetTokenHandler generates a new password reset token.
func (app *application) createPasswordResetTokenHandler(ctx *gin.Context) {
	var req createPasswordResetTokenRequest

	// Parse request body
	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate request body
	violations := validateCreatePasswordResetTokenRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Try to retrieve the corresponding user record for the email address.
	user, err := app.store.GetUserByEmail(ctx, *req.Email)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			violations.AddError("email", "no matching email address found")
			app.failedValidationResponse(ctx, violations)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// Return an error if the user account has not been activated yet.
	if !user.Activated {
		violations.AddError("email", "user account must be activated")
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Otherwise, create a new password reset token with a 45-minute expiry time.
	tokenPlaintext, _, err := app.store.GenerateToken(ctx, db.GenerateTokenParams{
		UserID:   user.ID,
		Duration: 45 * time.Minute,
		Scope:    db.ScopePasswordReset,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// Send an email to the user with the password reset token.
	app.background(func() {
		header := mailer.EmailHeader{
			Subject: "Reset your Greenlight password",
			To:      []string{user.Email},
		}

		data := map[string]any{
			"passwordResetToken": tokenPlaintext,
		}

		err := app.mailer.SendEmail(header, data, "token_password_rest.html")
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	// Send a 202 Accepted response and confirmation message to the client.
	rsp := envelope{"message": "an email will be sent to you containing password reset instructions"}
	app.writeJSON(ctx, http.StatusAccepted, rsp, nil)
}

type createActivationTokenRequest struct {
	Email *string `json:"email"`
}

func validateCreateActivationTokenRequest(req *createActivationTokenRequest) validator.Violations {
	violations := validator.New()

	if req.Email == nil {
		violations.AddError("email", "must be provided")
	} else if err := validator.ValidateUserEmail(*req.Email); err != nil {
		violations.AddError("email", err.Error())
	}

	return violations
}

// createActivationTokenHandler generates a new activation token for user.
func (app *application) createActivationTokenHandler(ctx *gin.Context) {
	var req createActivationTokenRequest

	// Parse the request body
	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	// Validate request body
	violations := validateCreateActivationTokenRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Try to retrieve the corresponding user record for the email address.
	user, err := app.store.GetUserByEmail(ctx, *req.Email)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			violations.AddError("email", "no matching email address found")
			app.failedValidationResponse(ctx, violations)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	// Return an error if the user has already been activated.
	if user.Activated {
		violations.AddError("email", "user has already been activated")
		app.failedValidationResponse(ctx, violations)
		return
	}

	// Otherwise, create a new activation token with a 3-day expiry time.
	tokenPlaintext, _, err := app.store.GenerateToken(ctx, db.GenerateTokenParams{
		UserID:   user.ID,
		Duration: 3 * 24 * time.Hour,
		Scope:    db.ScopeActivation,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	// Send an email to the user with the activation token.
	app.background(func() {
		header := mailer.EmailHeader{
			Subject: "Activate your Greenlight account",
			To:      []string{user.Email},
		}

		data := map[string]string{
			"activationToken": tokenPlaintext,
		}

		err := app.mailer.SendEmail(header, data, "token_activation.html")
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	// Send a 202 Accepted response and confirmation message to the client.
	rsp := envelope{"message": "an email will be sent to you containing activation instructions"}
	app.writeJSON(ctx, http.StatusAccepted, rsp, nil)
}
