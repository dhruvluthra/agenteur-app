package services

import (
	"context"
	"fmt"
	"strings"

	"agenteur.ai/api/internal/auth/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type pgxUserRepository struct{}

func NewUserRepository() types.UserRepository {
	return &pgxUserRepository{}
}

func (r *pgxUserRepository) Create(ctx context.Context, db database.DBTX, params types.CreateUserParams) (*types.User, error) {
	var u types.User
	err := db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at`,
		params.Email, params.PasswordHash, params.FirstName, params.LastName,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepository) GetByID(ctx context.Context, db database.DBTX, id uuid.UUID) (*types.User, error) {
	var u types.User
	err := db.QueryRow(ctx,
		`SELECT id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepository) GetByEmail(ctx context.Context, db database.DBTX, email string) (*types.User, error) {
	var u types.User
	err := db.QueryRow(ctx,
		`SELECT id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at
		 FROM users WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepository) Update(ctx context.Context, db database.DBTX, id uuid.UUID, params types.UpdateUserParams) (*types.User, error) {
	var u types.User
	err := db.QueryRow(ctx,
		`UPDATE users SET first_name = $2, last_name = $3, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at`,
		id, params.FirstName, params.LastName,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &u, nil
}

func (r *pgxUserRepository) ListAll(ctx context.Context, db database.DBTX, page, perPage int, search string) ([]*types.User, int, error) {
	offset := (page - 1) * perPage

	var total int
	var countQuery string
	var countArgs []any

	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		countQuery = `SELECT COUNT(*) FROM users WHERE LOWER(email) LIKE $1 OR LOWER(first_name) LIKE $1 OR LOWER(last_name) LIKE $1`
		countArgs = []any{like}
	} else {
		countQuery = `SELECT COUNT(*) FROM users`
	}

	if err := db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	var query string
	var args []any
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = `SELECT id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at
		         FROM users WHERE LOWER(email) LIKE $1 OR LOWER(first_name) LIKE $1 OR LOWER(last_name) LIKE $1
		         ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []any{like, perPage, offset}
	} else {
		query = `SELECT id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at
		         FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []any{perPage, offset}
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*types.User
	for rows.Next() {
		var u types.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}
	return users, total, nil
}

func (r *pgxUserRepository) SetSuperadmin(ctx context.Context, db database.DBTX, id uuid.UUID, isSuperadmin bool) (*types.User, error) {
	var u types.User
	err := db.QueryRow(ctx,
		`UPDATE users SET is_superadmin = $2, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, email, password_hash, first_name, last_name, is_superadmin, created_at, updated_at`,
		id, isSuperadmin,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("set superadmin: %w", err)
	}
	return &u, nil
}
