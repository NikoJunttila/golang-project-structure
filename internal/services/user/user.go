package userservice

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
)

// FetchUserWithEmail return user from email address
func FetchUserWithEmail(ctx context.Context, email string) (db.User, error) {
	user, err := db.Get().GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error(ctx, err, "Failed to FetchUserWithEmail")
		return db.User{}, err
	}
	return user, nil
}

// CheckUserExists tries to find user, retun true if finds user
func CheckUserExists(ctx context.Context, email string) (bool, error) {
	_, err := db.Get().GetUserByEmail(ctx, email)
	if err == nil {
		return true, nil // user exists
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // user does not exist
	}
	return false, err // actual error
}

// CreateUser add user to database either from oath or email
func CreateUser(ctx context.Context, password string, params CreateUserParams, oAuth OauthCreate) (db.User, error) {
	var passHash string
	var err error
	if !oAuth.IsOAuth {
		oAuth.Provider = params.Service
		//use password to hash
		passHash, err = auth.HashPassword(password)
		if err != nil {
			return db.User{}, err
		}
	}
	createParams := db.CreateUserParams{
		ID:            uuid.New().String(),
		LookupID:      uuid.New().String(),
		Email:         params.Email,
		PasswordHash:  passHash, // OAuth users don't have passwords
		Name:          params.Name,
		AvatarUrl:     params.AvatarURL,
		Provider:      oAuth.Provider,
		ProviderID:    oAuth.ProviderID,
		EmailVerified: oAuth.EmailVerified,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	newUser, err := db.Get().CreateUser(ctx, createParams)
	if err != nil {
		return db.User{}, err
	}

	return newUser, nil
}
