package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agenteur.ai/api/internal/auth/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailExists         = errors.New("email already registered")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type AuthService struct {
	pool            *pgxpool.Pool
	userRepo        types.UserRepository
	tokenRepo       types.RefreshTokenRepository
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	bcryptCost      int
}

func NewAuthService(
	pool *pgxpool.Pool,
	userRepo types.UserRepository,
	tokenRepo types.RefreshTokenRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		pool:            pool,
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		bcryptCost:      bcryptCost,
	}
}

// Signup creates a new user and returns the user, raw refresh token, and access JWT.
func (s *AuthService) Signup(ctx context.Context, email, password, firstName, lastName string) (*types.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existing, err := s.userRepo.GetByEmail(ctx, s.pool, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, "", "", ErrEmailExists
	}

	hash, err := HashPassword(password, s.bcryptCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, s.pool, types.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("create user: %w", err)
	}

	rawRefresh, accessJWT, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, rawRefresh, accessJWT, nil
}

// Login authenticates a user and returns the user, raw refresh token, and access JWT.
func (s *AuthService) Login(ctx context.Context, email, password string) (*types.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, s.pool, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	rawRefresh, accessJWT, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, rawRefresh, accessJWT, nil
}

// Logout deletes all refresh tokens for the user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.DeleteAllByUser(ctx, s.pool, userID)
}

// Refresh rotates refresh tokens and issues a new access JWT.
func (s *AuthService) Refresh(ctx context.Context, refreshTokenRaw string) (*types.User, string, string, error) {
	hash := HashToken(refreshTokenRaw)
	storedToken, err := s.tokenRepo.GetByHash(ctx, s.pool, hash)
	if err != nil {
		return nil, "", "", fmt.Errorf("get refresh token: %w", err)
	}
	if storedToken == nil || time.Now().After(storedToken.ExpiresAt) {
		return nil, "", "", ErrInvalidRefreshToken
	}

	if err := s.tokenRepo.DeleteByHash(ctx, s.pool, hash); err != nil {
		return nil, "", "", fmt.Errorf("delete old refresh token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, s.pool, storedToken.UserID)
	if err != nil {
		return nil, "", "", fmt.Errorf("get user for refresh: %w", err)
	}
	if user == nil {
		return nil, "", "", ErrInvalidRefreshToken
	}

	rawRefresh, accessJWT, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, rawRefresh, accessJWT, nil
}

func (s *AuthService) generateTokens(ctx context.Context, user *types.User) (string, string, error) {
	rawRefresh, refreshHash, err := GenerateRandomToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	_, err = s.tokenRepo.Create(ctx, s.pool, user.ID, refreshHash, time.Now().Add(s.refreshTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}

	accessJWT, err := GenerateAccessToken(user.ID, user.Email, user.IsSuperadmin, s.jwtSecret, s.accessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	return rawRefresh, accessJWT, nil
}
