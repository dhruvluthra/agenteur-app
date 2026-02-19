package services

import (
	"context"
	"fmt"

	"agenteur.ai/api/internal/auth/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool     *pgxpool.Pool
	userRepo types.UserRepository
}

func NewUserService(pool *pgxpool.Pool, userRepo types.UserRepository) *UserService {
	return &UserService{pool: pool, userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*types.User, error) {
	user, err := s.userRepo.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, params types.UpdateUserParams) (*types.User, error) {
	user, err := s.userRepo.Update(ctx, s.pool, id, params)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}
