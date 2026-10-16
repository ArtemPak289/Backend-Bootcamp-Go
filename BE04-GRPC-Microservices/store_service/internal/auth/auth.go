package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"store_api/internal/client"
	authpb "store_api/proto"
)

type contextKey string

const (
	AuthAttributeKey contextKey = "auth_attribute"
	UserClaimsKey    contextKey = "user_claims"
)

// AuthAttribute represents a self-written authorization attribute attached to controller methods.
type AuthAttribute struct {
	Required bool
	Roles    []string
}

// Authorize returns an AuthAttribute requiring user authentication.
func Authorize(roles ...string) AuthAttribute {
	return AuthAttribute{
		Required: true,
		Roles:    roles,
	}
}

// Public returns an AuthAttribute designating an endpoint as publicly accessible.
func Public() AuthAttribute {
	return AuthAttribute{
		Required: false,
	}
}

// WithAuth is a route wrapper that attaches the AuthAttribute to the request context.
func WithAuth(attr AuthAttribute, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), AuthAttributeKey, attr)
		handler(w, r.WithContext(ctx))
	}
}

// AuthMiddleware processes authentication and authorization via gRPC Auth Service.
type AuthMiddleware struct {
	authClient client.IAuthClient
}

func NewAuthMiddleware(authClient client.IAuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Whitelist public system endpoints
		if isPublicSystemPath(path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check self-written AuthAttribute attached to handler
		attr, ok := r.Context().Value(AuthAttributeKey).(AuthAttribute)
		if ok && !attr.Required {
			next.ServeHTTP(w, r)
			return
		}

		// Public authentication actions
		if isPublicAuthAction(path) {
			next.ServeHTTP(w, r)
			return
		}

		// All other API endpoints require gRPC verification
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "Missing Authorization header")
			return
		}

		// Extract token (supports both 'Bearer <token>' and raw token)
		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
		if token == "" {
			writeUnauthorized(w, "Empty or invalid token format in Authorization header")
			return
		}

		// Go to authorization service via gRPC to validate the token
		resp, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil || resp == nil || !resp.Valid {
			errMsg := "Invalid or expired token"
			if resp != nil && resp.ErrorMessage != "" {
				errMsg = resp.ErrorMessage
			}
			writeUnauthorized(w, errMsg)
			return
		}

		// Token is valid! Inject user claims into request context and forward to handler
		ctx := context.WithValue(r.Context(), UserClaimsKey, resp)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "Unauthorized",
		"message": message,
	})
}

func isPublicSystemPath(path string) bool {
	return path == "/health" ||
		path == "/swagger" ||
		path == "/swagger/" ||
		strings.HasPrefix(path, "/swagger/") ||
		path == "/api" ||
		path == "/api/" ||
		path == "/api/v1" ||
		path == "/api/v1/"
}

func isPublicAuthAction(path string) bool {
	return path == "/register" ||
		path == "/api/v1/register" ||
		path == "/auth" ||
		path == "/api/v1/auth" ||
		path == "/reset" ||
		path == "/api/v1/reset" ||
		path == "/change-password" ||
		path == "/api/v1/change-password"
}

// GetUserClaims returns the validated user claims from request context.
func GetUserClaims(r *http.Request) *authpb.ValidateTokenResponse {
	if val := r.Context().Value(UserClaimsKey); val != nil {
		if claims, ok := val.(*authpb.ValidateTokenResponse); ok {
			return claims
		}
	}
	return nil
}
