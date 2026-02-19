package types

import (
	"context"
	"time"

	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
)

// UserRepository defines user data access methods.
type UserRepository interface {
	Create(ctx context.Context, db database.DBTX, params CreateUserParams) (*User, error)
	GetByID(ctx context.Context, db database.DBTX, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, db database.DBTX, email string) (*User, error)
	Update(ctx context.Context, db database.DBTX, id uuid.UUID, params UpdateUserParams) (*User, error)
	ListAll(ctx context.Context, db database.DBTX, page, perPage int, search string) ([]*User, int, error)
	SetSuperadmin(ctx context.Context, db database.DBTX, id uuid.UUID, isSuperadmin bool) (*User, error)
}

// RefreshTokenRepository defines refresh token data access methods.
type RefreshTokenRepository interface {
	Create(ctx context.Context, db database.DBTX, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*RefreshToken, error)
	GetByHash(ctx context.Context, db database.DBTX, hash string) (*RefreshToken, error)
	DeleteByHash(ctx context.Context, db database.DBTX, hash string) error
	DeleteAllByUser(ctx context.Context, db database.DBTX, userID uuid.UUID) error
}
