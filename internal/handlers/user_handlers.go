package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	userService "github.com/nikojunttila/community/internal/services/user"
)

type createUserParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
type loginParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func PostCreateUserHandlerEmail(w http.ResponseWriter, r *http.Request) {
	var params createUserParams
	if !DecodeJSONBody(w, r, &params, 0) {
		fmt.Println("returned here")
		return
	}
	// Validation
	if params.Email == "" || params.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "email and password are required", userService.ErrParamsMismatch)
		return
	}
	if len(params.Password) < 8 {
		RespondWithError(w, http.StatusBadRequest, "password needs to be atleast 8 characters", userService.ErrTooWeakPassword)
		return
	}
	exists, err := userService.CheckUserExists(r.Context(), params.Email)
	if exists {
		RespondWithError(w, http.StatusBadRequest, "User already exists", userService.ErrUserAlreadyExists)
		return
	}
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}
	cleanedEmail := strings.TrimSpace(strings.ToLower(params.Email))
	createParams := userService.CreateUserParams{
		Email:   cleanedEmail,
		Name:    "",
		Service: string(userService.GetServiceEnumName(userService.Email)),
	}
	user, err := userService.CreateUser(r.Context(), params.Password, createParams, userService.OauthCreate{})
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error creating user: %v", err), err)
	}
	fmt.Println(user)
	RespondWithJson(w, http.StatusOK, "created user")
}

func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	params := loginParams{}
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}
	if params.Email == "" || params.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "no email or password provided", userService.ErrParamsMismatch)
		return
	}
	user, err := db.Get().GetUserByEmail(r.Context(), params.Email)
	if err == sql.ErrNoRows {
		RespondWithError(w, http.StatusBadRequest, "user with this email does not exist", userService.ErrUserNotFound)
		return
	}
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("database error %v", err), err)
		return
	}
	if user.Provider != string(userService.GetServiceEnumName(userService.Email)) {
		RespondWithError(w, http.StatusBadRequest, "error logging with email. user created using different authentication method", userService.ErrIncorrectAuthType)
		return
	}
	if !auth.CheckPasswordHash(params.Password, user.PasswordHash) {
		RespondWithError(w, http.StatusBadRequest, "wrong password or email", userService.ErrWrongPassword)
		return
	}
	token := auth.MakeToken(user.LookupID)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Path:  "/",
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: token,
	})
	RespondWithJson(w, http.StatusOK, token)
}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		fmt.Println(err)
		RespondWithError(w, 400, "error try again later", err)
	}
	fmt.Println(user)

	RespondWithJson(w, 200, user)
}
