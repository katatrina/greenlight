package db

import "context"

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type RegisterUserTxParams struct {
	Name           string
	Email          string
	HashedPassword []byte
	Permissions    []string
}

func (store *SQLStore) RegisterUserTx(ctx context.Context, arg RegisterUserTxParams) (User, error) {
	var user User

	/* Call execTx to execute a new transaction.
	The callback function will include all the queries that need to be executed within the transaction.
	Think of it as a blueprint of the transaction, where the actual execution will be done by the execTx method.
	*/
	err := store.execTx(ctx, func(qtx *Queries) error {
		var err error

		// Create a new user record.
		user, err = qtx.CreateUser(ctx, CreateUserParams{
			Name:           arg.Name,
			Email:          arg.Email,
			HashedPassword: arg.HashedPassword,
			Activated:      false,
		})
		if err != nil {
			return err
		}

		// Add permissions for the new user.
		err = qtx.AddPermissionsForUser(ctx, AddPermissionsForUserParams{
			UserID:          user.ID,
			PermissionCodes: arg.Permissions,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return user, err
}

func (store *SQLStore) ActivateUserTx(ctx context.Context, arg ActivateUserParams) (User, error) {
	var activatedUser User

	err := store.execTx(ctx, func(qtx *Queries) error {
		var err error

		// Set the user's activated field to true.
		activatedUser, err = qtx.ActivateUser(ctx, ActivateUserParams{
			UserID:  arg.UserID,
			Version: arg.Version,
		})
		if err != nil {
			return err
		}

		// If everything went successfully, then we delete all activation tokens of the user.
		err = qtx.DeleteUserTokens(ctx, DeleteUserTokensParams{
			UserID: arg.UserID,
			Scope:  ScopeActivation,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return activatedUser, err
}

type ResetUserPasswordTxParams struct {
	UserID         int64
	HashedPassword []byte
	Version        int32
}

func (store *SQLStore) ResetUserPasswordTx(ctx context.Context, arg ResetUserPasswordTxParams) error {
	return store.execTx(ctx, func(qtx *Queries) error {
		var err error

		// Set the user's password to the new hashed password.
		err = qtx.UpdateUserPassword(ctx, UpdateUserPasswordParams{
			UserID:         arg.UserID,
			HashedPassword: arg.HashedPassword,
			Version:        arg.Version,
		})
		if err != nil {
			return err
		}

		// If everything went successfully, then we delete all password reset tokens of the user.
		// We do this for security reasons.
		err = qtx.DeleteUserTokens(ctx, DeleteUserTokensParams{
			UserID: arg.UserID,
			Scope:  ScopePasswordReset,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
