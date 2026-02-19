package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"agenteur.ai/api/internal/auth/services"
	"agenteur.ai/api/internal/auth/types"
	"agenteur.ai/api/internal/httputil"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type updateUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r.Context())
	if claims == nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
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

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r.Context())
	if claims == nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	user, err := h.userService.Update(r.Context(), claims.UserID, types.UpdateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
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
