package types

import (
	"context"

	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
)

// OrganizationRepository defines org data access methods.
type OrganizationRepository interface {
	Create(ctx context.Context, db database.DBTX, params CreateOrgParams) (*Organization, error)
	GetByID(ctx context.Context, db database.DBTX, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, db database.DBTX, slug string) (*Organization, error)
	Update(ctx context.Context, db database.DBTX, id uuid.UUID, params UpdateOrgParams) (*Organization, error)
	ListAll(ctx context.Context, db database.DBTX) ([]*Organization, error)
}

// MembershipRepository defines org membership data access methods.
type MembershipRepository interface {
	Create(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID, role string) (*OrgMembership, error)
	GetByUserAndOrg(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID) (*OrgMembership, error)
	ListByUser(ctx context.Context, db database.DBTX, userID uuid.UUID) ([]*OrgWithRole, error)
	ListByOrg(ctx context.Context, db database.DBTX, orgID uuid.UUID) ([]*MemberWithUser, error)
	Delete(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID) error
	CountAdmins(ctx context.Context, db database.DBTX, orgID uuid.UUID) (int, error)
}

// InvitationRepository defines invitation data access methods.
type InvitationRepository interface {
	Create(ctx context.Context, db database.DBTX, params CreateInvitationParams) (*Invitation, error)
	GetByTokenHash(ctx context.Context, db database.DBTX, hash string) (*InvitationWithOrg, error)
	GetPendingByEmailAndOrg(ctx context.Context, db database.DBTX, email string, orgID uuid.UUID) (*Invitation, error)
	UpdateStatus(ctx context.Context, db database.DBTX, id uuid.UUID, status string) error
}

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendInvitation(ctx context.Context, to, inviterName, orgName, inviteURL string) error
}
