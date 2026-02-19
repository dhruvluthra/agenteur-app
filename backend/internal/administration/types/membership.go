package types

import (
	"time"

	"github.com/google/uuid"
)

type OrgMembership struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"userId"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// MemberWithUser is a join of org_memberships + users for member listing.
type MemberWithUser struct {
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joinedAt"`
}

// OrgWithRole is a join of organizations + org_memberships for user's org listing.
type OrgWithRole struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}
