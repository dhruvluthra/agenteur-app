package services

import (
	"context"
	"fmt"

	"agenteur.ai/api/internal/administration/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type pgxInvitationRepository struct{}

func NewInvitationRepository() types.InvitationRepository {
	return &pgxInvitationRepository{}
}

func (r *pgxInvitationRepository) Create(ctx context.Context, db database.DBTX, params types.CreateInvitationParams) (*types.Invitation, error) {
	var inv types.Invitation
	err := db.QueryRow(ctx,
		`INSERT INTO invitations (organization_id, invited_by, email, token_hash, role, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, organization_id, invited_by, email, token_hash, role, status, expires_at, created_at, updated_at`,
		params.OrganizationID, params.InvitedBy, params.Email, params.TokenHash, params.Role, params.ExpiresAt,
	).Scan(&inv.ID, &inv.OrganizationID, &inv.InvitedBy, &inv.Email, &inv.TokenHash, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}
	return &inv, nil
}

func (r *pgxInvitationRepository) GetByTokenHash(ctx context.Context, db database.DBTX, hash string) (*types.InvitationWithOrg, error) {
	var inv types.InvitationWithOrg
	err := db.QueryRow(ctx,
		`SELECT i.id, i.organization_id, i.invited_by, i.email, i.token_hash, i.role, i.status, i.expires_at, i.created_at, i.updated_at,
		        o.name, COALESCE(u.first_name || ' ' || u.last_name, u.email)
		 FROM invitations i
		 JOIN organizations o ON o.id = i.organization_id
		 JOIN users u ON u.id = i.invited_by
		 WHERE i.token_hash = $1`, hash,
	).Scan(&inv.ID, &inv.OrganizationID, &inv.InvitedBy, &inv.Email, &inv.TokenHash, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UpdatedAt,
		&inv.OrganizationName, &inv.InvitedByName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get invitation by token hash: %w", err)
	}
	return &inv, nil
}

func (r *pgxInvitationRepository) GetPendingByEmailAndOrg(ctx context.Context, db database.DBTX, email string, orgID uuid.UUID) (*types.Invitation, error) {
	var inv types.Invitation
	err := db.QueryRow(ctx,
		`SELECT id, organization_id, invited_by, email, token_hash, role, status, expires_at, created_at, updated_at
		 FROM invitations
		 WHERE LOWER(email) = LOWER($1) AND organization_id = $2 AND status = 'pending'`,
		email, orgID,
	).Scan(&inv.ID, &inv.OrganizationID, &inv.InvitedBy, &inv.Email, &inv.TokenHash, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get pending invitation: %w", err)
	}
	return &inv, nil
}

func (r *pgxInvitationRepository) UpdateStatus(ctx context.Context, db database.DBTX, id uuid.UUID, status string) error {
	_, err := db.Exec(ctx,
		`UPDATE invitations SET status = $2, updated_at = NOW() WHERE id = $1`,
		id, status)
	if err != nil {
		return fmt.Errorf("update invitation status: %w", err)
	}
	return nil
}
