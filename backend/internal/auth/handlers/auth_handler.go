package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"agenteur.ai/api/internal/auth/services"
	"agenteur.ai/api/internal/httputil"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthHandler struct {
	authService    *services.AuthService
	jwtSecret      string
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
	secureCookies  bool
}

func NewAuthHandler(authService *services.AuthService, jwtSecret string, accessTTL, refreshTTL time.Duration, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		secureCookies:   secureCookies,
	}
}

type signupRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	IsSuperadmin bool   `json:"isSuperadmin"`
	CreatedAt    string `json:"createdAt"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	errs := make(map[string]string)
	if req.Email == "" || !emailRegex.MatchString(req.Email) {
		errs["email"] = "Valid email is required"
	}
	if len(req.Password) < 8 {
		errs["password"] = "Password must be at least 8 characters"
	}
	if req.FirstName == "" {
		errs["firstName"] = "First name is required"
	}
	if len(errs) > 0 {
		httputil.ValidationError(w, "Validation failed", errs)
		return
	}

	user, rawRefresh, accessJWT, err := h.authService.Signup(r.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, services.ErrEmailExists) {
			httputil.Error(w, http.StatusConflict, "CONFLICT", "Email already registered")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	h.setAuthCookies(w, accessJWT, rawRefresh)
	httputil.JSON(w, http.StatusCreated, userResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsSuperadmin: user.IsSuperadmin,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	user, rawRefresh, accessJWT, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	h.setAuthCookies(w, accessJWT, rawRefresh)
	httputil.JSON(w, http.StatusOK, userResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsSuperadmin: user.IsSuperadmin,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired refresh token")
		return
	}

	user, rawRefresh, accessJWT, err := h.authService.Refresh(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired refresh token")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	h.setAuthCookies(w, accessJWT, rawRefresh)
	httputil.JSON(w, http.StatusOK, userResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsSuperadmin: user.IsSuperadmin,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	claims, err := services.ParseAccessTokenUnvalidated(cookie.Value, h.jwtSecret)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	if err := h.authService.Logout(r.Context(), claims.UserID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	h.clearAuthCookies(w)
	httputil.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(h.accessTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/auth",
		MaxAge:   int(h.refreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}
