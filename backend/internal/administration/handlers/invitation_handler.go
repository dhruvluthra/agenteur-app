package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"agenteur.ai/api/internal/administration/services"
	authhandlers "agenteur.ai/api/internal/auth/handlers"
	"agenteur.ai/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InvitationHandler struct {
	invitationService *services.InvitationService
	orgService        *services.OrgService
}

func NewInvitationHandler(invitationService *services.InvitationService, orgService *services.OrgService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		orgService:        orgService,
	}
}

type createInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type invitationResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expiresAt"`
	CreatedAt string `json:"createdAt"`
}

type invitationDetailResponse struct {
	OrganizationName string `json:"organizationName"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	InvitedByName    string `json:"invitedByName"`
	ExpiresAt        string `json:"expiresAt"`
}

func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	if req.Email == "" {
		httputil.ValidationError(w, "Validation failed", map[string]string{"email": "Email is required"})
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Role != "admin" && req.Role != "user" {
		httputil.ValidationError(w, "Validation failed", map[string]string{"role": "Role must be 'admin' or 'user'"})
		return
	}

	// Get org name and inviter name for the email
	org, err := h.orgService.Get(r.Context(), orgID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	inv, err := h.invitationService.Create(r.Context(), orgID, claims.UserID, req.Email, req.Role, claims.Email, org.Name)
	if err != nil {
		if errors.Is(err, services.ErrInvitationExists) {
			httputil.Error(w, http.StatusConflict, "CONFLICT", "Invitation already pending for this email")
			return
		}
		if errors.Is(err, services.ErrAlreadyMember) {
			httputil.Error(w, http.StatusConflict, "CONFLICT", "User is already a member")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusCreated, invitationResponse{
		ID:        inv.ID.String(),
		Email:     inv.Email,
		Role:      inv.Role,
		Status:    inv.Status,
		ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
		CreatedAt: inv.CreatedAt.Format(time.RFC3339),
	})
}

func (h *InvitationHandler) GetByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		httputil.Error(w, http.StatusBadRequest, "INVALID_TOKEN", "Token is required")
		return
	}

	inv, err := h.invitationService.GetByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, services.ErrInvitationNotFound) {
			httputil.Error(w, http.StatusNotFound, "NOT_FOUND", "Invitation not found or expired")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusOK, invitationDetailResponse{
		OrganizationName: inv.OrganizationName,
		Email:            inv.Email,
		Role:             inv.Role,
		InvitedByName:    inv.InvitedByName,
		ExpiresAt:        inv.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	claims := authhandlers.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		httputil.Error(w, http.StatusBadRequest, "INVALID_TOKEN", "Token is required")
		return
	}

	membership, err := h.invitationService.Accept(r.Context(), token, claims.UserID, claims.Email)
	if err != nil {
		if errors.Is(err, services.ErrInvitationNotFound) {
			httputil.Error(w, http.StatusNotFound, "NOT_FOUND", "Invitation not found or expired")
			return
		}
		if errors.Is(err, services.ErrEmailMismatch) {
			httputil.Error(w, http.StatusForbidden, "FORBIDDEN", "Email does not match invitation")
			return
		}
		if errors.Is(err, services.ErrAlreadyMember) {
			httputil.Error(w, http.StatusConflict, "CONFLICT", "Already a member of this organization")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]any{
		"membership": map[string]string{
			"userId":         membership.UserID.String(),
			"organizationId": membership.OrganizationID.String(),
			"role":           membership.Role,
			"joinedAt":       membership.CreatedAt.Format(time.RFC3339),
		},
	})
}
