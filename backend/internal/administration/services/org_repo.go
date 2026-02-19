package services

import (
	"context"
	"fmt"

	"agenteur.ai/api/internal/administration/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type pgxOrganizationRepository struct{}

func NewOrganizationRepository() types.OrganizationRepository {
	return &pgxOrganizationRepository{}
}

func (r *pgxOrganizationRepository) Create(ctx context.Context, db database.DBTX, params types.CreateOrgParams) (*types.Organization, error) {
	var o types.Organization
	err := db.QueryRow(ctx,
		`INSERT INTO organizations (name, slug)
		 VALUES ($1, $2)
		 RETURNING id, name, slug, created_at, updated_at`,
		params.Name, params.Slug,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	return &o, nil
}

func (r *pgxOrganizationRepository) GetByID(ctx context.Context, db database.DBTX, id uuid.UUID) (*types.Organization, error) {
	var o types.Organization
	err := db.QueryRow(ctx,
		`SELECT id, name, slug, created_at, updated_at FROM organizations WHERE id = $1`, id,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get organization by id: %w", err)
	}
	return &o, nil
}

func (r *pgxOrganizationRepository) GetBySlug(ctx context.Context, db database.DBTX, slug string) (*types.Organization, error) {
	var o types.Organization
	err := db.QueryRow(ctx,
		`SELECT id, name, slug, created_at, updated_at FROM organizations WHERE LOWER(slug) = LOWER($1)`, slug,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get organization by slug: %w", err)
	}
	return &o, nil
}

func (r *pgxOrganizationRepository) Update(ctx context.Context, db database.DBTX, id uuid.UUID, params types.UpdateOrgParams) (*types.Organization, error) {
	var o types.Organization
	err := db.QueryRow(ctx,
		`UPDATE organizations SET name = $2, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, name, slug, created_at, updated_at`,
		id, params.Name,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update organization: %w", err)
	}
	return &o, nil
}

func (r *pgxOrganizationRepository) ListAll(ctx context.Context, db database.DBTX) ([]*types.Organization, error) {
	rows, err := db.Query(ctx,
		`SELECT id, name, slug, created_at, updated_at FROM organizations ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*types.Organization
	for rows.Next() {
		var o types.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		orgs = append(orgs, &o)
	}
	return orgs, nil
}
