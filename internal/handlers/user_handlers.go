package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/cache"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
	userService "github.com/nikojunttila/community/internal/services/user"
)

var (
	// emailRegex is used to validate email addresses using a basic regular expression.
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// ErrInvalidEmail indicates an email did not match the required format.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrWeakPassword indicates a password does not meet minimum strength requirements.
	ErrWeakPassword = errors.New("password must be at least 8 characters")

	// ErrMissingFields indicates required fields were not provided in the request.
	ErrMissingFields = errors.New("email and password are required")
)

// CreateUserRequest represents the JSON payload for registering a user via email.
type CreateUserRequest struct {
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
}

// LoginRequest represents the JSON payload for user login.
type LoginRequest struct {
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

// LoginResponse represents the successful login response containing token and user.
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user,omitempty"`
}

// User represents the essential user data returned in responses.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// validateCreateUserRequest validates and sanitizes input for user creation.
func validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Email == "" || req.Password == "" {
		return ErrMissingFields
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !emailRegex.MatchString(req.Email) {
		return ErrInvalidEmail
	}

	if len(req.Password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

// validateLoginRequest validates input fields for user login.
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

// PostCreateUserHandlerEmail handles user registration via email form submission.
func PostCreateUserHandlerEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateUserRequest
	err := r.ParseForm()
	if err != nil {
		RespondWithError(ctx, w, http.StatusBadRequest, "invalid form data", err)
		return
	}

	req.Email = r.FormValue("email")
	req.Password = r.FormValue("password")

	// Validate input
	if err := validateCreateUserRequest(&req); err != nil {
		var statusCode int
		var serviceErr error

		switch err {
		case ErrMissingFields, ErrInvalidEmail:
			statusCode = http.StatusBadRequest
			serviceErr = userService.ErrParamsMismatch
		case ErrWeakPassword:
			statusCode = http.StatusBadRequest
			serviceErr = userService.ErrTooWeakPassword
		default:
			statusCode = http.StatusBadRequest
			serviceErr = err
		}
		RespondWithError(ctx, w, statusCode, err.Error(), serviceErr)
		return
	}

	// Check if user already exists
	exists, err := userService.CheckUserExists(ctx, req.Email)
	if err != nil {
		logger.Error(ctx, err, "failed to check user existence")
		RespondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	if exists {
		RespondWithError(ctx, w, http.StatusConflict, "User already exists", userService.ErrUserAlreadyExists)
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
		RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}
	RespondWithJSON(ctx, w, http.StatusCreated, map[string]string{
		"message": "User created successfully",
		"userID":  user.ID,
	})
}

// PostLoginHandler handles user login and returns a JWT token upon success.
func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req LoginRequest
	if !DecodeJSONBody(w, r, &req, 0) {
		return
	}

	// Validate input
	if err := validateLoginRequest(&req); err != nil {
		RespondWithError(ctx, w, http.StatusBadRequest, err.Error(), userService.ErrParamsMismatch)
		return
	}

	dbUser, err := db.Get().GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(ctx, w, http.StatusBadRequest, "Invalid email or password", userService.ErrWrongPassword)
			return
		}
		RespondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	if dbUser.Provider != string(userService.GetServiceEnumName(userService.Email)) {
		RespondWithError(ctx, w, http.StatusBadRequest,
			"Please use the authentication method you originally signed up with",
			userService.ErrIncorrectAuthType)
		return
	}

	if !auth.CheckPasswordHash(req.Password, dbUser.PasswordHash) {
		RespondWithError(ctx, w, http.StatusBadRequest, "Invalid email or password", userService.ErrWrongPassword)
		return
	}

	// Generate JWT token
	token := auth.MakeToken(dbUser.LookupID, dbUser.Role)

	// Set secure cookie with JWT
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
	RespondWithJSON(ctx, w, http.StatusOK, response)
}

// GetProfileHandler returns the authenticated user's profile based on the context.
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := cache.GetUser(ctx)
	if err != nil {
		RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to find active user", err)
		return
	}
	RespondWithJSON(ctx, w, http.StatusOK, user)
}

// GetCreatePage renders the HTML page for creating a new user via form.
func GetCreatePage(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "createUser.html", nil)
	if err != nil {
		RespondWithError(r.Context(), w, http.StatusInternalServerError, "internal server error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

