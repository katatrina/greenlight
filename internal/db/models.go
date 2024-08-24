// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"time"
)

type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Runtime     Runtime   `json:"runtime"`
	Genres      []string  `json:"genres"`
	PublishYear int32     `json:"publish_year"`
	Version     int32     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
}

type Permission struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
}

type Token struct {
	UserID    int64     `json:"user_id"`
	Hash      []byte    `json:"hash"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword []byte    `json:"-"`
	Activated      bool      `json:"activated"`
	Version        int32     `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

type UserPermission struct {
	UserID       int64 `json:"user_id"`
	PermissionID int64 `json:"permission_id"`
}
