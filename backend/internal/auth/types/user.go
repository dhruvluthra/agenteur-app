package types

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	IsSuperadmin bool      `json:"isSuperadmin"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

type UpdateUserParams struct {
	FirstName string
	LastName  string
}
