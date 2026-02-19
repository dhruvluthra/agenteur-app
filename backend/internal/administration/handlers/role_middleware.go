package handlers

import (
	"context"
	"net/http"

	admintypes "agenteur.ai/api/internal/administration/types"
	authhandlers "agenteur.ai/api/internal/auth/handlers"
	authtypes "agenteur.ai/api/internal/auth/types"
	"agenteur.ai/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roleContextKey string

const (
	membershipKey roleContextKey = "org_membership"
)

// RoleMiddleware provides org-level and superadmin authorization checks.
type RoleMiddleware struct {
	pool           *pgxpool.Pool
	membershipRepo admintypes.MembershipRepository
	userRepo       authtypes.UserRepository
}

func NewRoleMiddleware(pool *pgxpool.Pool, membershipRepo admintypes.MembershipRepository, userRepo authtypes.UserRepository) *RoleMiddleware {
	return &RoleMiddleware{
		pool:           pool,
		membershipRepo: membershipRepo,
		userRepo:       userRepo,
	}
}

// RequireOrgMember checks that the authenticated user is a member of the org
// (from {orgID} URL param) or is a superadmin.
func (m *RoleMiddleware) RequireOrgMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := authhandlers.GetUserClaims(r.Context())
		if claims == nil {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}

		orgIDStr := chi.URLParam(r, "orgID")
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid organization ID")
			return
		}

		// Check if superadmin (from DB, not JWT)
		user, err := m.userRepo.GetByID(r.Context(), m.pool, claims.UserID)
		if err != nil || user == nil {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}
		if user.IsSuperadmin {
			ctx := context.WithValue(r.Context(), membershipKey, &admintypes.OrgMembership{
				UserID:         claims.UserID,
				OrganizationID: orgID,
				Role:           "admin",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		membership, err := m.membershipRepo.GetByUserAndOrg(r.Context(), m.pool, claims.UserID, orgID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}
		if membership == nil {
			httputil.Error(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to perform this action")
			return
		}

		ctx := context.WithValue(r.Context(), membershipKey, membership)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireOrgAdmin checks that the user has the admin role in the org or is a superadmin.
func (m *RoleMiddleware) RequireOrgAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		membership := GetOrgMembership(r.Context())
		if membership == nil {
			httputil.Error(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to perform this action")
			return
		}
		if membership.Role != "admin" {
			httputil.Error(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to perform this action")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireSuperadmin checks that the user is a superadmin (from DB).
func (m *RoleMiddleware) RequireSuperadmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := authhandlers.GetUserClaims(r.Context())
		if claims == nil {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}

		user, err := m.userRepo.GetByID(r.Context(), m.pool, claims.UserID)
		if err != nil || user == nil {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}
		if !user.IsSuperadmin {
			httputil.Error(w, http.StatusForbidden, "FORBIDDEN", "You don't have permission to perform this action")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetOrgMembership retrieves the org membership from context, set by RequireOrgMember.
func GetOrgMembership(ctx context.Context) *admintypes.OrgMembership {
	m, ok := ctx.Value(membershipKey).(*admintypes.OrgMembership)
	if !ok {
		return nil
	}
	return m
}
