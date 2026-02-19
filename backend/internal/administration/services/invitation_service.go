package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agenteur.ai/api/internal/administration/types"
	authservices "agenteur.ai/api/internal/auth/services"
	authtypes "agenteur.ai/api/internal/auth/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvitationExists   = errors.New("invitation already pending for this email")
	ErrAlreadyMember      = errors.New("user is already a member")
	ErrInvitationNotFound = errors.New("invitation not found or expired")
	ErrEmailMismatch      = errors.New("email does not match invitation")
)

type InvitationService struct {
	pool           *pgxpool.Pool
	invitationRepo types.InvitationRepository
	membershipRepo types.MembershipRepository
	emailService   types.EmailService
	userRepo       authtypes.UserRepository
	inviteBaseURL  string
	inviteTokenTTL time.Duration
}

func NewInvitationService(
	pool *pgxpool.Pool,
	invitationRepo types.InvitationRepository,
	membershipRepo types.MembershipRepository,
	emailService types.EmailService,
	userRepo authtypes.UserRepository,
	inviteBaseURL string,
	inviteTokenTTL time.Duration,
) *InvitationService {
	return &InvitationService{
		pool:           pool,
		invitationRepo: invitationRepo,
		membershipRepo: membershipRepo,
		emailService:   emailService,
		userRepo:       userRepo,
		inviteBaseURL:  inviteBaseURL,
		inviteTokenTTL: inviteTokenTTL,
	}
}

// Create creates a new invitation for an email to join an organization.
func (s *InvitationService) Create(ctx context.Context, orgID, invitedByUserID uuid.UUID, email, role, inviterName, orgName string) (*types.Invitation, error) {
	// Check if already a member
	existingUser, err := s.userRepo.GetByEmail(ctx, s.pool, email)
	if err != nil {
		return nil, fmt.Errorf("check user: %w", err)
	}
	if existingUser != nil {
		membership, err := s.membershipRepo.GetByUserAndOrg(ctx, s.pool, existingUser.ID, orgID)
		if err != nil {
			return nil, fmt.Errorf("check membership: %w", err)
		}
		if membership != nil {
			return nil, ErrAlreadyMember
		}
	}

	// Check for existing pending invite
	existing, err := s.invitationRepo.GetPendingByEmailAndOrg(ctx, s.pool, email, orgID)
	if err != nil {
		return nil, fmt.Errorf("check existing invite: %w", err)
	}
	if existing != nil {
		return nil, ErrInvitationExists
	}

	rawToken, tokenHash, err := authservices.GenerateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	inv, err := s.invitationRepo.Create(ctx, s.pool, types.CreateInvitationParams{
		OrganizationID: orgID,
		InvitedBy:      invitedByUserID,
		Email:          email,
		TokenHash:      tokenHash,
		Role:           role,
		ExpiresAt:      time.Now().Add(s.inviteTokenTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	inviteURL := s.inviteBaseURL + "/" + rawToken
	if err := s.emailService.SendInvitation(ctx, email, inviterName, orgName, inviteURL); err != nil {
		return nil, fmt.Errorf("send invitation email: %w", err)
	}

	return inv, nil
}

// GetByToken looks up an invitation by raw token for display.
func (s *InvitationService) GetByToken(ctx context.Context, rawToken string) (*types.InvitationWithOrg, error) {
	hash := authservices.HashToken(rawToken)
	inv, err := s.invitationRepo.GetByTokenHash(ctx, s.pool, hash)
	if err != nil {
		return nil, fmt.Errorf("get invitation: %w", err)
	}
	if inv == nil || inv.Status != "pending" || time.Now().After(inv.ExpiresAt) {
		return nil, ErrInvitationNotFound
	}
	return inv, nil
}

// Accept accepts an invitation and creates an org membership.
func (s *InvitationService) Accept(ctx context.Context, rawToken string, userID uuid.UUID, userEmail string) (*types.OrgMembership, error) {
	hash := authservices.HashToken(rawToken)

	var membership *types.OrgMembership
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		inv, err := s.invitationRepo.GetByTokenHash(ctx, tx, hash)
		if err != nil {
			return fmt.Errorf("get invitation: %w", err)
		}
		if inv == nil || inv.Status != "pending" || time.Now().After(inv.ExpiresAt) {
			return ErrInvitationNotFound
		}

		if !strings.EqualFold(userEmail, inv.Email) {
			return ErrEmailMismatch
		}

		// Check already a member
		existing, err := s.membershipRepo.GetByUserAndOrg(ctx, tx, userID, inv.OrganizationID)
		if err != nil {
			return fmt.Errorf("check membership: %w", err)
		}
		if existing != nil {
			return ErrAlreadyMember
		}

		membership, err = s.membershipRepo.Create(ctx, tx, userID, inv.OrganizationID, inv.Role)
		if err != nil {
			return fmt.Errorf("create membership: %w", err)
		}

		return s.invitationRepo.UpdateStatus(ctx, tx, inv.ID, "accepted")
	})
	if err != nil {
		return nil, err
	}
	return membership, nil
}
