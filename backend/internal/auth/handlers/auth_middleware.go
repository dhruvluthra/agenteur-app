package handlers

import (
	"context"
	"net/http"

	"agenteur.ai/api/internal/auth/services"
	"agenteur.ai/api/internal/httputil"
)

type contextKey string

const claimsKey contextKey = "user_claims"

// AuthMiddleware validates JWT access tokens from cookies.
type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

// Authenticate reads the access_token cookie, validates the JWT, and stores
// claims in context for downstream handlers.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}

		claims, err := services.ValidateAccessToken(cookie.Value, m.jwtSecret)
		if err != nil {
			httputil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims retrieves TokenClaims from context, set by Authenticate middleware.
func GetUserClaims(ctx context.Context) *services.TokenClaims {
	claims, ok := ctx.Value(claimsKey).(*services.TokenClaims)
	if !ok {
		return nil
	}
	return claims
}
