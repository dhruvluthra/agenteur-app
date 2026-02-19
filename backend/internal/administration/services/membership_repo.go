package services

import (
	"context"
	"fmt"

	"agenteur.ai/api/internal/administration/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type pgxMembershipRepository struct{}

func NewMembershipRepository() types.MembershipRepository {
	return &pgxMembershipRepository{}
}

func (r *pgxMembershipRepository) Create(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID, role string) (*types.OrgMembership, error) {
	var m types.OrgMembership
	err := db.QueryRow(ctx,
		`INSERT INTO org_memberships (user_id, organization_id, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, organization_id, role, created_at, updated_at`,
		userID, orgID, role,
	).Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create membership: %w", err)
	}
	return &m, nil
}

func (r *pgxMembershipRepository) GetByUserAndOrg(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID) (*types.OrgMembership, error) {
	var m types.OrgMembership
	err := db.QueryRow(ctx,
		`SELECT id, user_id, organization_id, role, created_at, updated_at
		 FROM org_memberships WHERE user_id = $1 AND organization_id = $2`,
		userID, orgID,
	).Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get membership: %w", err)
	}
	return &m, nil
}

func (r *pgxMembershipRepository) ListByUser(ctx context.Context, db database.DBTX, userID uuid.UUID) ([]*types.OrgWithRole, error) {
	rows, err := db.Query(ctx,
		`SELECT o.id, o.name, o.slug, m.role, o.created_at
		 FROM org_memberships m
		 JOIN organizations o ON o.id = m.organization_id
		 WHERE m.user_id = $1
		 ORDER BY o.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list orgs by user: %w", err)
	}
	defer rows.Close()

	var orgs []*types.OrgWithRole
	for rows.Next() {
		var o types.OrgWithRole
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.Role, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan org with role: %w", err)
		}
		orgs = append(orgs, &o)
	}
	return orgs, nil
}

func (r *pgxMembershipRepository) ListByOrg(ctx context.Context, db database.DBTX, orgID uuid.UUID) ([]*types.MemberWithUser, error) {
	rows, err := db.Query(ctx,
		`SELECT u.id, u.email, u.first_name, u.last_name, m.role, m.created_at
		 FROM org_memberships m
		 JOIN users u ON u.id = m.user_id
		 WHERE m.organization_id = $1
		 ORDER BY m.created_at ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members by org: %w", err)
	}
	defer rows.Close()

	var members []*types.MemberWithUser
	for rows.Next() {
		var m types.MemberWithUser
		if err := rows.Scan(&m.UserID, &m.Email, &m.FirstName, &m.LastName, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *pgxMembershipRepository) Delete(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`DELETE FROM org_memberships WHERE user_id = $1 AND organization_id = $2`,
		userID, orgID)
	if err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

func (r *pgxMembershipRepository) CountAdmins(ctx context.Context, db database.DBTX, orgID uuid.UUID) (int, error) {
	var count int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM org_memberships WHERE organization_id = $1 AND role = 'admin'`,
		orgID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return count, nil
}
