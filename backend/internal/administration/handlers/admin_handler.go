package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	authservices "agenteur.ai/api/internal/auth/services"
	"agenteur.ai/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminHandler struct {
	pool     *pgxpool.Pool
	userService *authservices.UserService
}

func NewAdminHandler(pool *pgxpool.Pool, userService *authservices.UserService) *AdminHandler {
	return &AdminHandler{pool: pool, userService: userService}
}

type userResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	IsSuperadmin bool   `json:"isSuperadmin"`
	CreatedAt    string `json:"createdAt"`
}

type toggleSuperadminRequest struct {
	IsSuperadmin bool `json:"isSuperadmin"`
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	search := r.URL.Query().Get("search")

	users, total, err := h.userService.ListAll(r.Context(), page, perPage, search)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	resp := make([]userResponse, len(users))
	for i, u := range users {
		resp[i] = userResponse{
			ID:           u.ID.String(),
			Email:        u.Email,
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			IsSuperadmin: u.IsSuperadmin,
			CreatedAt:    u.CreatedAt.Format(time.RFC3339),
		}
	}

	httputil.JSON(w, http.StatusOK, map[string]any{
		"users":   resp,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}

func (h *AdminHandler) ToggleSuperadmin(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req toggleSuperadminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	user, err := h.userService.SetSuperadmin(r.Context(), userID, req.IsSuperadmin)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	if user == nil {
		httputil.Error(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	httputil.JSON(w, http.StatusOK, userResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsSuperadmin: user.IsSuperadmin,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}
