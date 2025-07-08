
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/markbates/goth/gothic"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
	userService "github.com/nikojunttila/community/internal/services/user"
)

// GetAuthCallBack handles the OAuth callback after the user authenticates with a third-party provider.
// It completes the auth process, creates the user if they don't exist, sets a JWT cookie, and redirects to the frontend.
func GetAuthCallBack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		// Print session values to help debug missing cookies or invalid session state
		session, _ := gothic.Store.Get(r, "gothic_session")
		fmt.Printf("Session values: %+v\n", session.Values)
		RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to complete user auth", err)
		return
	}

	// Check if the user already exists in the database
	exists, err := userService.CheckUserExists(r.Context(), userInfo.Email)
	if err != nil {
		RespondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	var user db.User
	if exists {
		user, err = db.Get().GetUserByEmail(r.Context(), userInfo.Email)
		if err != nil {
			RespondWithError(ctx, w, http.StatusInternalServerError, "Internal server error", err)
			return
		}
	} else {
		// Register new user
		user, err = userService.CreateUser(r.Context(), "", userService.CreateUserParams{
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			AvatarURL: userInfo.AvatarURL,
		}, userService.OauthCreate{
			IsOAuth:       true,
			EmailVerified: true,
			Provider:      userInfo.Provider,
			ProviderID:    userInfo.UserID,
		})
		if err != nil {
			RespondWithError(ctx, w, http.StatusInternalServerError, "Failed to create user", err)
			return
		}
	}

	// Create and set JWT cookie
	jwtToken := auth.MakeToken(user.LookupID, user.Role)
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt", // Must be "jwt" to be recognized by jwtauth.Verifier
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		// Secure: true, // Uncomment in production over HTTPS
	})

	logger.Info(ctx, fmt.Sprintf("User authenticated: %+v\n", user))

	// Redirect user to frontend
	http.Redirect(w, r, "http://localhost:3000/", http.StatusFound)
}

// GetBeginAuth starts the OAuth authentication process for a given provider.
// The frontend should call this endpoint to begin the login flow (e.g., /public/google/begin).
func GetBeginAuth(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

