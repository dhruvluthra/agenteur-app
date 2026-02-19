package services

import (
	"context"
	"fmt"
	"time"

	"agenteur.ai/api/internal/auth/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type pgxRefreshTokenRepository struct{}

func NewRefreshTokenRepository() types.RefreshTokenRepository {
	return &pgxRefreshTokenRepository{}
}

func (r *pgxRefreshTokenRepository) Create(ctx context.Context, db database.DBTX, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*types.RefreshToken, error) {
	var t types.RefreshToken
	err := db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, created_at`,
		userID, tokenHash, expiresAt,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	return &t, nil
}

func (r *pgxRefreshTokenRepository) GetByHash(ctx context.Context, db database.DBTX, hash string) (*types.RefreshToken, error) {
	var t types.RefreshToken
	err := db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}
	return &t, nil
}

func (r *pgxRefreshTokenRepository) DeleteByHash(ctx context.Context, db database.DBTX, hash string) error {
	_, err := db.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, hash)
	if err != nil {
		return fmt.Errorf("delete refresh token by hash: %w", err)
	}
	return nil
}

func (r *pgxRefreshTokenRepository) DeleteAllByUser(ctx context.Context, db database.DBTX, userID uuid.UUID) error {
	_, err := db.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete all refresh tokens by user: %w", err)
	}
	return nil
}
