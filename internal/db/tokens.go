package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password-reset"
)

type GenerateTokenParams struct {
	UserID   int64
	Duration time.Duration
	Scope    string
}

func (store *SQLStore) GenerateToken(ctx context.Context, arg GenerateTokenParams) (tokenPlaintext string, token Token, err error) {
	randomBytes := make([]byte, 16)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", token, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token
	// Plaintext field. This will be the token string that we send to the user in their
	// welcome email. They will look similar to this:
	//
	// Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	//
	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	tokenPlaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	// Generate a SHA-256 hash of the plaintext token string. This will be the value
	// that we store in the `hash` field of our database table. Note that the
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using the [:] operator before storing it.
	hash := sha256.Sum256([]byte(tokenPlaintext))

	token, err = store.CreateToken(ctx, CreateTokenParams{
		Hash:      hash[:],
		UserID:    arg.UserID,
		ExpiresAt: time.Now().Add(arg.Duration),
		Scope:     arg.Scope,
	})
	if err != nil {
		return "", token, err
	}

	return tokenPlaintext, token, nil
}
