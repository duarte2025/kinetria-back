package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	domainauth "github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// AuthHandler handles HTTP requests for auth endpoints.
type AuthHandler struct {
	registerUC     *domainauth.RegisterUC
	loginUC        *domainauth.LoginUC
	refreshTokenUC *domainauth.RefreshTokenUC
	logoutUC       *domainauth.LogoutUC
	jwtManager     *gatewayauth.JWTManager
	validate       *validator.Validate
}

// NewAuthHandler creates a new AuthHandler with all required dependencies.
func NewAuthHandler(
	registerUC *domainauth.RegisterUC,
	loginUC *domainauth.LoginUC,
	refreshTokenUC *domainauth.RefreshTokenUC,
	logoutUC *domainauth.LogoutUC,
	jwtManager *gatewayauth.JWTManager,
	validate *validator.Validate,
) *AuthHandler {
	return &AuthHandler{
		registerUC:     registerUC,
		loginUC:        loginUC,
		refreshTokenUC: refreshTokenUC,
		logoutUC:       logoutUC,
		jwtManager:     jwtManager,
		validate:       validate,
	}
}

// Register handles POST /auth/register
// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} SuccessResponse{data=AuthResponse}
// @Failure 409 {object} ErrorResponse "Email already exists"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var req struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	output, err := h.registerUC.Execute(r.Context(), domainauth.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "An account with this email already exists.")
		case errors.Is(err, domainerrors.ErrMalformedParameters):
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Password must be at least 8 characters.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}
	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"accessToken":  output.AccessToken,
		"refreshToken": output.RefreshToken,
		"expiresIn":    output.ExpiresIn,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} SuccessResponse{data=AuthResponse}
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	output, err := h.loginUC.Execute(r.Context(), domainauth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect.")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}
	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"accessToken":  output.AccessToken,
		"refreshToken": output.RefreshToken,
		"expiresIn":    output.ExpiresIn,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} SuccessResponse{data=RefreshResponse}
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	output, err := h.refreshTokenUC.Execute(r.Context(), domainauth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainerrors.ErrTokenInvalid),
			errors.Is(err, domainerrors.ErrTokenExpired),
			errors.Is(err, domainerrors.ErrTokenRevoked):
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}
	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"accessToken":  output.AccessToken,
		"refreshToken": output.RefreshToken,
		"expiresIn":    output.ExpiresIn,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to invalidate"
// @Success 200 {object} SuccessResponse{data=MessageResponse}
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	// Validate JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if _, err := h.jwtManager.ParseToken(tokenStr); err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		return
	}

	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	if _, err := h.logoutUC.Execute(r.Context(), domainauth.LogoutInput{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": code, "message": message})
}
