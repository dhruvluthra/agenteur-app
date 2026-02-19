package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"agenteur.ai/api/internal/administration/types"
	"agenteur.ai/api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLastAdmin  = errors.New("cannot remove the last admin from the organization")
	ErrNotFound   = errors.New("resource not found")
	ErrMemberNotFound = errors.New("member not found")
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

type OrgService struct {
	pool           *pgxpool.Pool
	orgRepo        types.OrganizationRepository
	membershipRepo types.MembershipRepository
}

func NewOrgService(pool *pgxpool.Pool, orgRepo types.OrganizationRepository, membershipRepo types.MembershipRepository) *OrgService {
	return &OrgService{
		pool:           pool,
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
	}
}

// Create creates a new organization and assigns the user as admin.
func (s *OrgService) Create(ctx context.Context, userID uuid.UUID, name string) (*types.Organization, error) {
	slug := generateSlug(name)

	var org *types.Organization
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		// Check slug collision, append random suffix if needed
		existing, err := s.orgRepo.GetBySlug(ctx, tx, slug)
		if err != nil {
			return err
		}
		if existing != nil {
			suffix, err := randomSuffix()
			if err != nil {
				return err
			}
			slug = slug + "-" + suffix
		}

		org, err = s.orgRepo.Create(ctx, tx, types.CreateOrgParams{
			Name: name,
			Slug: slug,
		})
		if err != nil {
			return err
		}

		_, err = s.membershipRepo.Create(ctx, tx, userID, org.ID, "admin")
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("create org: %w", err)
	}
	return org, nil
}

// List returns the user's orgs, or all orgs if superadmin.
func (s *OrgService) List(ctx context.Context, userID uuid.UUID, isSuperadmin bool) ([]*types.OrgWithRole, error) {
	if isSuperadmin {
		orgs, err := s.orgRepo.ListAll(ctx, s.pool)
		if err != nil {
			return nil, err
		}
		result := make([]*types.OrgWithRole, len(orgs))
		for i, o := range orgs {
			result[i] = &types.OrgWithRole{
				ID:        o.ID,
				Name:      o.Name,
				Slug:      o.Slug,
				Role:      "admin",
				CreatedAt: o.CreatedAt,
			}
		}
		return result, nil
	}
	return s.membershipRepo.ListByUser(ctx, s.pool, userID)
}

// Get returns an organization by ID.
func (s *OrgService) Get(ctx context.Context, orgID uuid.UUID) (*types.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, s.pool, orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrNotFound
	}
	return org, nil
}

// Update updates an organization's name.
func (s *OrgService) Update(ctx context.Context, orgID uuid.UUID, name string) (*types.Organization, error) {
	return s.orgRepo.Update(ctx, s.pool, orgID, types.UpdateOrgParams{Name: name})
}

// ListMembers returns all members of an organization.
func (s *OrgService) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*types.MemberWithUser, error) {
	return s.membershipRepo.ListByOrg(ctx, s.pool, orgID)
}

// RemoveMember removes a member from an organization.
func (s *OrgService) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	membership, err := s.membershipRepo.GetByUserAndOrg(ctx, s.pool, userID, orgID)
	if err != nil {
		return err
	}
	if membership == nil {
		return ErrMemberNotFound
	}

	if membership.Role == "admin" {
		count, err := s.membershipRepo.CountAdmins(ctx, s.pool, orgID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastAdmin
		}
	}

	return s.membershipRepo.Delete(ctx, s.pool, userID, orgID)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "org"
	}
	return slug
}

func randomSuffix() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
