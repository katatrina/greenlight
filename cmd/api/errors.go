package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/validator"
)

// logError is a generic helper for logging an error (usually caused by the internal server)
// along with the current request method and URL as attributes in the log entry.
func (app *application) logError(ctx *gin.Context, err error) {
	var (
		method = ctx.Request.Method
		uri    = ctx.Request.URL
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// errorResponse is a generic helper for sending JSON-formatted error
// messages to the client with a given status code.
func (app *application) errorResponse(ctx *gin.Context, statusCode int, message any) {
	rsp := envelope{"error": message}

	app.writeJSON(ctx, statusCode, rsp, nil)
}

// notFoundResponse send a 404 Not Found status code and
// JSON response to the client.
func (app *application) notFoundResponse(ctx *gin.Context) {
	message := "the requested resource could not be found"

	app.errorResponse(ctx, http.StatusNotFound, message)
}

// serverErrorResponse will be used when the application encounters an
// unexpected problem at runtime.
//
// It logs the detailed error message, then sends a 500 Internal Server Error status code
// and JSON response (containing a generic error message) to the client.
func (app *application) serverErrorResponse(ctx *gin.Context, err error) {
	app.logError(ctx, err)

	message := "the server encountered a problem and could not process your request"

	app.errorResponse(ctx, http.StatusInternalServerError, message)
}

// methodNotAllowedResponse send a 405 Method Not Allowed status code
// and JSON response to the client.
func (app *application) methodNotAllowedResponse(ctx *gin.Context) {
	message := fmt.Sprintf("the %s method is not supported for this resource", ctx.Request.Method)

	app.errorResponse(ctx, http.StatusMethodNotAllowed, message)
}

// badRequestResponse send a 400 Bad Request status code
// and JSON response to the client.
func (app *application) badRequestResponse(ctx *gin.Context, err error) {
	app.errorResponse(ctx, http.StatusBadRequest, err.Error())
}

// failedValidationResponse send a 422 Unprocessable Entity status code and JSON response to the client.
func (app *application) failedValidationResponse(ctx *gin.Context, violations validator.Violations) {
	app.errorResponse(ctx, http.StatusUnprocessableEntity, violations)
}

// editConflictResponse send a 409 Conflict status code and JSON response to the client.
func (app *application) editConflictResponse(ctx *gin.Context) {
	message := "unable to update the record due to an edit conflict, please try again"

	app.errorResponse(ctx, http.StatusConflict, message)
}

// integrityConstraintViolationResponse send a 409 Conflict status code and JSON response to the client.
func (app *application) integrityConstraintViolationResponse(ctx *gin.Context, message string) {
	app.errorResponse(ctx, http.StatusConflict, message)
}

// invalidCrendentialsReponse send 401 Unauthorized status code and a generic error message to the client.
func (app *application) invalidCredentialsResponse(ctx *gin.Context) {
	message := "invalid authentication credentials"

	app.errorResponse(ctx, http.StatusUnauthorized, message)
}

// invalidAuthenticationTokenResponse send 401 Unauthorized status code and a generic error message to the client.
func (app *application) invalidAuthenticationTokenResponse(ctx *gin.Context) {
	ctx.Header("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"

	app.errorResponse(ctx, http.StatusUnauthorized, message)
}

// authenticatedUserRequiredResponse send 401 Unauthorized status code and a generic error message to the client.
func (app *application) authenticatedUserRequiredResponse(ctx *gin.Context) {
	message := "you must be an authenticated user in order to access this resource"

	app.errorResponse(ctx, http.StatusUnauthorized, message)
}

// activatedAccountRequiredResponse send 403 Forbidden status code and a generic error message to the client.
func (app *application) activatedAccountRequiredResponse(ctx *gin.Context) {
	message := "your user account must be activated to access this resource"

	app.errorResponse(ctx, http.StatusForbidden, message)
}

// notPermittedResponse send 403 Forbidden status code and a generic error message to the client.
func (app *application) notPermittedResponse(ctx *gin.Context) {
	message := "your user account doesn't have the necessary permissions to access this resource"

	app.errorResponse(ctx, http.StatusForbidden, message)
}

// mismatchedAuthenticatedUserEmailResponse sends 403 Forbidden status code and a generic error message to the client.
func (app *application) mismatchedAuthenticatedUserEmailResponse(ctx *gin.Context) {
	message := "the email provided does not match the authenticated user's email"

	app.errorResponse(ctx, http.StatusForbidden, message)
}
