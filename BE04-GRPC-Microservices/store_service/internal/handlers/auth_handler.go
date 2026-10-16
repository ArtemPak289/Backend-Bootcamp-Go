package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"store_api/internal/client"
)

type AuthHandler struct {
	authClient client.IAuthClient
}

func NewAuthHandler(authClient client.IAuthClient) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

type RegisterRequestDTO struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResetPasswordRequestDTO struct {
	Email string `json:"email"`
}

type ChangePasswordRequestDTO struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Register creates a new user via gRPC Auth Service and returns signed JWT token.
// @Summary Register a new user
// @Description Register a new user in the appliance store using Auth Service via gRPC
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequestDTO true "User registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		respondError(w, http.StatusBadRequest, "email, password, first_name and last_name are required")
		return
	}

	res, err := h.authClient.Register(r.Context(), req.Email, req.FirstName, req.LastName, req.PhoneNumber, req.Password)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Registration failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"token":        res.GetToken(),
		"user_id":      res.GetUserId(),
		"email":        res.GetEmail(),
		"first_name":   res.GetFirstName(),
		"last_name":    res.GetLastName(),
		"phone_number": res.GetPhoneNumber(),
	})
}

// Login authenticates user via gRPC Auth Service and returns signed JWT token.
// @Summary User authentication
// @Description Authenticate user credentials against Auth Service via gRPC
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequestDTO true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	res, err := h.authClient.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Authentication failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token":        res.GetToken(),
		"user_id":      res.GetUserId(),
		"email":        res.GetEmail(),
		"first_name":   res.GetFirstName(),
		"last_name":    res.GetLastName(),
		"phone_number": res.GetPhoneNumber(),
	})
}

// ResetPassword requests a password recovery via gRPC Auth Service.
// @Summary Reset password
// @Description Trigger password recovery; outputs temporary password to Auth Service console
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequestDTO true "Password reset request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Router /reset [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	res, err := h.authClient.ResetPassword(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Password reset failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.GetSuccess(),
		"message": res.GetMessage(),
	})
}

// ChangePassword updates user password via gRPC Auth Service.
// @Summary Change password
// @Description Change existing user password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequestDTO true "Password change request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Router /change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.OldPassword == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "email, old_password, and new_password are required")
		return
	}

	res, err := h.authClient.ChangePassword(r.Context(), req.Email, req.OldPassword, req.NewPassword)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Password change failed: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.GetSuccess(),
		"message": res.GetMessage(),
	})
}
