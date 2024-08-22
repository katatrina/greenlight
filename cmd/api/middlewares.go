package main

import (
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katatrina/greenlight/internal/db"
	"github.com/katatrina/greenlight/internal/validator"
)

// authenticate indicates which user a request is coming from.
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
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
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
