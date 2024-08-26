package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/util"
	"github.com/katatrina/greenlight/internal/validator"
)

type createAuthenticationTokenRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type createAuthenticationTokenResponse struct {
	TokenPlaintext string    `json:"token"`
	ExpiredAt      time.Time `json:"expired_at"`
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

	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

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
			ExpiredAt:      token.ExpiresAt,
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

	if err := app.readJSON(ctx, &req); err != nil {
		app.badRequestResponse(ctx, err)
		return
	}

	violations := validateCreatePasswordResetTokenRequest(&req)
	if !violations.Empty() {
		app.failedValidationResponse(ctx, violations)
		return
	}

	user, err := app.store.GetUserByEmail(ctx, *req.Email)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			violations.AddError("email", "no matching email address")
			app.failedValidationResponse(ctx, violations)
			return
		}

		app.serverErrorResponse(ctx, err)
		return
	}

	if !user.Activated {
		violations.AddError("email", "user account must be activated")
		app.failedValidationResponse(ctx, violations)
		return
	}

	tokenPlaintext, _, err := app.store.GenerateToken(ctx, db.GenerateTokenParams{
		UserID:   user.ID,
		Duration: 45 * time.Minute,
		Scope:    db.ScopePasswordReset,
	})
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"passwordResetToken": tokenPlaintext,
		}

		err = app.mailer.SendEmail("Change Greenlight password", data, []string{user.Email}, nil, nil, nil, "token_password_rest.html")
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	rsp := envelope{"message": "an email will be sent to you containing password reset instructions"}
	app.writeJSON(ctx, http.StatusAccepted, rsp, nil)
}
