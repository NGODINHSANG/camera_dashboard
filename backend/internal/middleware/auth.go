package middleware

import (
	"context"
	"net/http"
	"strings"

	"camera-dashboard-backend/pkg/jwt"
	"camera-dashboard-backend/pkg/response"
)

type contextKey string

const UserContextKey contextKey = "user"

// Claims is an alias for jwt.Claims for easier access from handlers
type Claims = jwt.Claims

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Authorization header required")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.Unauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwt.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				if err == jwt.ErrExpiredToken {
					response.Unauthorized(w, "Token has expired")
					return
				}
				response.Unauthorized(w, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims == nil {
			response.Unauthorized(w, "Authentication required")
			return
		}

		if claims.Role != "admin" {
			response.Forbidden(w, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) *jwt.Claims {
	claims, ok := ctx.Value(UserContextKey).(*jwt.Claims)
	if !ok {
		return nil
	}
	return claims
}
