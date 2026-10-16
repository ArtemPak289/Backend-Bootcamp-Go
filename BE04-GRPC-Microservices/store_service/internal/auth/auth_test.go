package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authpb "store_api/proto"
)

type mockAuthClient struct {
	validToken string
}

func (m *mockAuthClient) Register(ctx context.Context, email, firstName, lastName, phone, password string) (*authpb.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthClient) Login(ctx context.Context, email, password string) (*authpb.AuthResponse, error) {
	return nil, nil
}
func (m *mockAuthClient) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) (*authpb.ChangePasswordResponse, error) {
	return nil, nil
}
func (m *mockAuthClient) ResetPassword(ctx context.Context, email string) (*authpb.ResetPasswordResponse, error) {
	return nil, nil
}
func (m *mockAuthClient) ValidateToken(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error) {
	if token == m.validToken {
		return &authpb.ValidateTokenResponse{
			Valid:     true,
			UserId:    "123",
			Email:     "user@example.com",
			FirstName: "Jane",
			LastName:  "Doe",
		}, nil
	}
	return &authpb.ValidateTokenResponse{
		Valid:        false,
		ErrorMessage: "invalid token",
	}, nil
}
func (m *mockAuthClient) GetUser(ctx context.Context, userID, email string) (*authpb.UserResponse, error) {
	return nil, nil
}
func (m *mockAuthClient) Close() error {
	return nil
}

func TestAuthMiddleware(t *testing.T) {
	mockClient := &mockAuthClient{validToken: "good-token"}
	middleware := NewAuthMiddleware(mockClient)

	// Protected handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil {
			t.Error("Claims should not be nil in protected handler")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := middleware.Middleware(nextHandler)

	// 1. Missing Authorization header -> 401
	req1 := httptest.NewRequest("GET", "/api/v1/clients", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for missing auth header, got %d", rec1.Code)
	}

	// 2. Invalid Token -> 401
	req2 := httptest.NewRequest("GET", "/api/v1/clients", nil)
	req2.Header.Set("Authorization", "Bearer bad-token")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for invalid token, got %d", rec2.Code)
	}

	// 3. Valid Token -> 200
	req3 := httptest.NewRequest("GET", "/api/v1/clients", nil)
	req3.Header.Set("Authorization", "Bearer good-token")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("Expected 200 for valid token, got %d", rec3.Code)
	}

	// 4. Public endpoint without token -> 200
	publicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	decoratedPublicHandler := WithAuth(Public(), publicHandler)
	wrappedPublic := middleware.Middleware(decoratedPublicHandler)

	req4 := httptest.NewRequest("POST", "/register", nil)
	rec4 := httptest.NewRecorder()
	wrappedPublic.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("Expected 200 for public endpoint without token, got %d", rec4.Code)
	}
}
