package types

import (
	"time"

	"github.com/google/uuid"
)

type Invitation struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	InvitedBy      uuid.UUID `json:"invitedBy"`
	Email          string    `json:"email"`
	TokenHash      string    `json:"-"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expiresAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// InvitationWithOrg is a join with org name and inviter name for display.
type InvitationWithOrg struct {
	Invitation
	OrganizationName string `json:"organizationName"`
	InvitedByName    string `json:"invitedByName"`
}

type CreateInvitationParams struct {
	OrganizationID uuid.UUID
	InvitedBy      uuid.UUID
	Email          string
	TokenHash      string
	Role           string
	ExpiresAt      time.Time
}
