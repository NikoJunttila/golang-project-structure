package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/logger"
	"github.com/nikojunttila/community/internal/services/email"
	userS "github.com/nikojunttila/community/internal/services/user"
)

// GetProfileHandler retrieves the authenticated user's profile
func GetProfileHandlerAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Info(ctx,"admin profile access")
	user, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		RespondWithError(w,ctx, http.StatusInternalServerError, "Failed to find active user", err)
		return
	}
	go func(){
		email.Mailer.Send(context.Background(),"","nikosamulijunttila@gmail.com","better test","<h1>hellope!</h1>","hellope")
	}()
	RespondWithJson(w,ctx, http.StatusOK, user)
}

// GetProfileHandler retrieves the authenticated user's profile
func GetProfileAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	admin, err := auth.GetUserFromContext(ctx)
	if err != nil {
		RespondWithError(w,r.Context() ,http.StatusInternalServerError, "Failed to find active user", err)
		return
	}
	type userFetchParams struct {
		Email string `json:"email"`
	}
	var params userFetchParams
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}
	user, err := userS.FetchUserWithEmail(ctx, params.Email)
	if err != nil {
		RespondWithError(w,ctx ,http.StatusBadRequest, "Could not fetch user with this email", err)
		return
	}

	logger.Info(r.Context(), fmt.Sprintf("Profile accessed %s by %s", user.ID, admin.ID))
	RespondWithJson(w,ctx, http.StatusOK, user)
}
