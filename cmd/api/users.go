package main

import (
	"net/http"

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
	} else if err := validator.ValidateUserPassword(*req.Password); err != nil {
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

	err = app.mailer.SendEmail("Welcome <3!!!", user, []string{user.Email}, nil, nil, nil, "user_welcome.html")
	if err != nil {
		app.serverErrorResponse(ctx, err)
		return
	}

	rsp := envelop{"user": user}
	app.writeJSON(ctx, http.StatusCreated, rsp, nil)
}
