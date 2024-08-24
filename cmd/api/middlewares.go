package main

import (
	"crypto/sha256"
	"errors"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/validator"
)

// authenticate middleware indicates which user a request is coming from, either an authenticated user or an anonymous user.
func (app *application) authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		ctx.Header("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This will
		// return an empty string "" if there is no such header found.
		authorizationHeader := ctx.GetHeader("Authorization")

		// If there is no Authorization header found, use the contextSetUser() helper
		// that we just made to add the AnonymousUser to the request context. Then we
		// call the next handler in the chain and return without executing any of the
		// code below.
		if authorizationHeader == "" {
			app.contextSetUser(ctx, db.AnonymousUser)
			ctx.Next()
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into its constituent parts, and if the
		// header isn't in the expected format we return a 401 Unauthorized response.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" || headerParts[1] == "" {
			app.invalidAuthenticationTokenResponse(ctx)
			ctx.Abort()
			return
		}

		// Extract the actual authentication token from the header parts.
		token := headerParts[1]

		// Validate the token to make sure it is in a sensible format.
		err := validator.ValidateTokenPlaintext(token)
		if err != nil {
			app.invalidAuthenticationTokenResponse(ctx)
			ctx.Abort()
			return
		}

		tokenHash := sha256.Sum256([]byte(token))

		// Retrieve the details of the user associated with the authentication token,
		// again calling the invalidAuthenticationTokenResponse() helper if no
		// matching record was found.
		user, err := app.store.GetUserByToken(ctx, db.GetUserByTokenParams{
			Hash:  tokenHash[:],
			Scope: db.ScopeAuthentication,
		})
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				app.invalidAuthenticationTokenResponse(ctx)
				ctx.Abort()
				return
			}

			app.serverErrorResponse(ctx, err)
			return
		}

		// Add the user information to the request context.
		app.contextSetUser(ctx, &user)

		ctx.Next()
	}
}

// requireActivatedUser middleware restricts access to activated user accounts.
func (app *application) requireActivatedUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve the user information from the request context.
		user := app.contextGetUser(ctx)

		// If the user account is not activated, we inform them that they need to activate their account.
		if !user.Activated {
			app.activatedAccountRequiredResponse(ctx)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// requireAuthenticatedUser middleware restricts access to authenticated users.
func (app *application) requireAuthenticatedUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve the user information from the request context.
		user := app.contextGetUser(ctx)

		// If the user is anonymous, we inform the client that they should authenticate before trying again.
		if user.IsAnonymous() {
			app.authenticatedUserRequiredResponse(ctx)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// requirePermission middleware restricts access to users with the appropriate permission.
func (app *application) requirePermission(code string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve the user from the request context.
		user := app.contextGetUser(ctx)

		// Get a slice of permissions for the user.
		permissions, err := app.store.GetUserPermissions(ctx, user.ID)
		if err != nil {
			app.serverErrorResponse(ctx, err)
			ctx.Abort()
			return
		}

		// Check if the slice includes the required permission. If it doesn't, then
		// return a 403 Forbidden response.
		if !slices.Contains(permissions, code) {
			app.notPermittedResponse(ctx)
			ctx.Abort()
			return
		}

		// Otherwise they have the required permission so we call the next handler in the chain.
		ctx.Next()
	}
}
