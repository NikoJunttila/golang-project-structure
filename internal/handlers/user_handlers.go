package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	userService "github.com/nikojunttila/community/internal/services/user"
)

var (
	// Email validation regex (basic but more robust than none)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	// Common validation errors
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrWeakPassword    = errors.New("password must be at least 8 characters")
	ErrMissingFields   = errors.New("email and password are required")
)

type CreateUserRequest struct {
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"required,email"`
}

type LoginRequest struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// validateCreateUserRequest validates and sanitizes user creation input
func validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Email == "" || req.Password == "" {
		return ErrMissingFields
	}
	
	// Clean and validate email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !emailRegex.MatchString(req.Email) {
		return ErrInvalidEmail
	}
	
	if len(req.Password) < 8 {
		return ErrWeakPassword
	}
	
	return nil
}

// validateLoginRequest validates login input
func validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return ErrMissingFields
	}
	
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !emailRegex.MatchString(req.Email) {
		return ErrInvalidEmail
	}
	
	return nil
}

// PostCreateUserHandlerEmail handles user registration via email
func PostCreateUserHandlerEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req CreateUserRequest
	if !DecodeJSONBody(w, r, &req, 0) {
		log.Error().Msg("failed to decode request body")
		return
	}
	
	// Validate input
	if err := validateCreateUserRequest(&req); err != nil {
		var statusCode int
		var serviceErr error
		
		switch err {
		case ErrMissingFields:
			statusCode = http.StatusBadRequest
			serviceErr = userService.ErrParamsMismatch
		case ErrInvalidEmail:
			statusCode = http.StatusBadRequest
			serviceErr = userService.ErrParamsMismatch
		case ErrWeakPassword:
			statusCode = http.StatusBadRequest
			serviceErr = userService.ErrTooWeakPassword
		default:
			statusCode = http.StatusBadRequest
			serviceErr = err
		}
		
		RespondWithError(w, statusCode, err.Error(), serviceErr)
		return
	}
	
	// Check if user already exists
	exists, err := userService.CheckUserExists(ctx, req.Email)
	if err != nil {
		log.Error().Msgf("failed to check user existence %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	
	if exists {
		RespondWithError(w, http.StatusConflict, "User already exists", userService.ErrUserAlreadyExists)
		return
	}
	
	// Create user
	createParams := userService.CreateUserParams{
		Email:   req.Email,
		Name:    "",
		Service: string(userService.GetServiceEnumName(userService.Email)),
	}
	
	user, err := userService.CreateUser(ctx, req.Password, createParams, userService.OauthCreate{})
	if err != nil {
		slog.Error("failed to create user", "error", err, "email", req.Email)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}
	
	slog.Info("user created successfully", "userID", user.ID, "email", req.Email)
	RespondWithJson(w, http.StatusCreated, map[string]string{
		"message": "User created successfully",
		"userID":  user.ID,
	})
}

// PostLoginHandler handles user authentication
func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !DecodeJSONBody(w, r, &req, 0) {
		return
	}
	// Validate input
	if err := validateLoginRequest(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), userService.ErrParamsMismatch)
		return
	}
	dbUser, err := db.Get().GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Don't reveal whether user exists or password is wrong for security
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password", userService.ErrWrongPassword)
			return
		}
		
		log.Error().Msgf("database error during login %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	if dbUser.Provider != string(userService.GetServiceEnumName(userService.Email)) {
		RespondWithError(w, http.StatusBadRequest, 
			"Please use the authentication method you originally signed up with", 
			userService.ErrIncorrectAuthType)
		return
	}
	if !auth.CheckPasswordHash(req.Password, dbUser.PasswordHash) {
		slog.Warn("failed login attempt", "email", req.Email)
		RespondWithError(w, http.StatusUnauthorized, "Invalid email or password", userService.ErrWrongPassword)
		return
	}
	
	// Generate JWT token
	token := auth.MakeToken(dbUser.LookupID)
	
	// Set secure cookie
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		// Secure: true, // Enable in production with HTTPS
	}
	http.SetCookie(w, cookie)
	
	user := &User{
		ID:       dbUser.LookupID,
		Email:    dbUser.Email,
		Name:     dbUser.Name,
		Provider: dbUser.Provider,
	}
	
	response := LoginResponse{
		Token: token,
		User:  user,
	}
	
	log.Info().Msgf("successful login %s %s", dbUser.LookupID, req.Email)
	RespondWithJson(w, http.StatusOK, response)
}

// GetProfileHandler retrieves the authenticated user's profile
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		slog.Error("failed to get user from context", "error", err)
		RespondWithError(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}
	
	slog.Info("profile accessed", "userID", user.ID)
	RespondWithJson(w, http.StatusOK, user)
}
