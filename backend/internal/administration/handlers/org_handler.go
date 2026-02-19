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

type OrgHandler struct {
	orgService *services.OrgService
}

func NewOrgHandler(orgService *services.OrgService) *OrgHandler {
	return &OrgHandler{orgService: orgService}
}

type createOrgRequest struct {
	Name string `json:"name"`
}

type orgResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"createdAt"`
}

type memberResponse struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
	JoinedAt  string `json:"joinedAt"`
}

func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := authhandlers.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req createOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		httputil.ValidationError(w, "Validation failed", map[string]string{"name": "Name is required"})
		return
	}

	org, err := h.orgService.Create(r.Context(), claims.UserID, req.Name)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusCreated, orgResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: org.CreatedAt.Format(time.RFC3339),
	})
}

func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := authhandlers.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	orgs, err := h.orgService.List(r.Context(), claims.UserID, claims.IsSuperadmin)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := make([]orgResponse, len(orgs))
	for i, o := range orgs {
		resp[i] = orgResponse{
			ID:        o.ID.String(),
			Name:      o.Name,
			Slug:      o.Slug,
			CreatedAt: o.CreatedAt.Format(time.RFC3339),
		}
	}

	httputil.JSON(w, http.StatusOK, map[string]any{"organizations": resp})
}

func (h *OrgHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid organization ID")
		return
	}

	org, err := h.orgService.Get(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "NOT_FOUND", "Organization not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusOK, orgResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: org.CreatedAt.Format(time.RFC3339),
	})
}

func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid organization ID")
		return
	}

	var req createOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		httputil.ValidationError(w, "Validation failed", map[string]string{"name": "Name is required"})
		return
	}

	org, err := h.orgService.Update(r.Context(), orgID, req.Name)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusOK, orgResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: org.CreatedAt.Format(time.RFC3339),
	})
}

func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid organization ID")
		return
	}

	members, err := h.orgService.ListMembers(r.Context(), orgID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := make([]memberResponse, len(members))
	for i, m := range members {
		resp[i] = memberResponse{
			UserID:    m.UserID.String(),
			Email:     m.Email,
			FirstName: m.FirstName,
			LastName:  m.LastName,
			Role:      m.Role,
			JoinedAt:  m.JoinedAt.Format(time.RFC3339),
		}
	}

	httputil.JSON(w, http.StatusOK, map[string]any{"members": resp})
}

func (h *OrgHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid organization ID")
		return
	}

	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	err = h.orgService.RemoveMember(r.Context(), orgID, userID)
	if err != nil {
		if errors.Is(err, services.ErrLastAdmin) {
			httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cannot remove the last admin")
			return
		}
		if errors.Is(err, services.ErrMemberNotFound) {
			httputil.Error(w, http.StatusNotFound, "NOT_FOUND", "Member not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}
