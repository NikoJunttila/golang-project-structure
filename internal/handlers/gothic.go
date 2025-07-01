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

func GetAuthCallBack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		// Check if session exists
		session, _ := gothic.Store.Get(r, "gothic_session")
		fmt.Printf("Session values: %+v\n", session.Values)
		RespondWithError(w,ctx ,http.StatusInternalServerError, "Failed to complete user auth", err)
		return
	}
	exists, err := userService.CheckUserExists(r.Context(), userInfo.Email)
	if err != nil {
		RespondWithError(w,ctx, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	var user db.User
	if exists {
		user, err = db.Get().GetUserByEmail(r.Context(), userInfo.Email)
		if err != nil {
			RespondWithError(w,ctx, http.StatusInternalServerError, "Internal server error", err)
		}
	} else {
		user, err = userService.CreateUser(r.Context(), "", userService.CreateUserParams{
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			AvatarUrl: userInfo.AvatarURL,
		}, userService.OauthCreate{
			IsOAuth:       true,
			EmailVerified: true,
			Provider:      userInfo.Provider,
			ProviderID:    userInfo.UserID,
		})
		if err != nil {
			RespondWithError(w, ctx, http.StatusInternalServerError,"Failed to create user", err)
			return
		}
	}
	jwtToken := auth.MakeToken(user.LookupID, user.Role)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Path:  "/",
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: jwtToken,
	})

	logger.Info(ctx,fmt.Sprintf("User authenticated: %+v\n", user))
	http.Redirect(w, r, "http://localhost:3000/", http.StatusFound)
}
func GetBeginAuth(w http.ResponseWriter, r *http.Request) {
	//client just needs to have this function
	// async function loginWithgoogle(){
	// window.location.href = "/public/google/begin"}
	gothic.BeginAuthHandler(w, r)
}
