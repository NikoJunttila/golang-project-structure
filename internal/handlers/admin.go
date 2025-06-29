package handlers

import (
	"fmt"
	"net/http"

	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/logger"
	userS "github.com/nikojunttila/community/internal/services/user"
	"github.com/rs/zerolog/log"
)

// GetProfileHandler retrieves the authenticated user's profile
func GetProfileAdmin(w http.ResponseWriter, r *http.Request) {
	type userFetchParams struct {
		Email string `json:"email"`
	}
	var params userFetchParams
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}
	admin, err := auth.GetUserFromContext(r.Context())
	log.Info().Msgf("Email: %s", admin.Email)
	log.Error().Err(err).Msg("")
	fmt.Println(err)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to find current user", err)
		return
	}
	user, err := userS.FetchUserWithEmail(r.Context(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Could not fetch user with this email", err)
		return
	}

	logger.Info(fmt.Sprintf("Profile accessed %s by %s", user.ID, admin.ID))
	RespondWithJson(w, http.StatusOK, user)
}
