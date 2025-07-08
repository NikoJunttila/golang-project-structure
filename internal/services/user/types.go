// Package userservice contains necessary functionality for user
package userservice

import "errors"

// AuthServiceEnum represents the authentication method used by a user.
type AuthServiceEnum string

const (
	// Email represents authentication via email and password.
	Email AuthServiceEnum = "email"

	// Google represents authentication via Google OAuth.
	Google AuthServiceEnum = "google"

	// Discord represents authentication via Discord OAuth.
	Discord AuthServiceEnum = "discord"
)

// ErrUserAlreadyExists indicates that the user already exists.
var ErrUserAlreadyExists = errors.New("user already exists")

// ErrParamsMismatch indicates missing or incorrect parameters.
var ErrParamsMismatch = errors.New("necessary params not provided")

// ErrIncorrectAuthType indicates that the user is using a different authentication method.
var ErrIncorrectAuthType = errors.New("user already created using a different login method (google, discord etc...)")

// ErrWrongPassword indicates that the provided password is incorrect.
var ErrWrongPassword = errors.New("wrong password")

// ErrTooWeakPassword indicates that the provided password does not meet strength requirements.
var ErrTooWeakPassword = errors.New("weak password")

// GetServiceEnumName returns the given AuthServiceEnum as-is.
// Useful for type safety or validation logic.
func GetServiceEnumName(service AuthServiceEnum) AuthServiceEnum {
	return service
}

// OauthCreate contains information required when creating a user via OAuth.
type OauthCreate struct {
	IsOAuth       bool
	EmailVerified bool
	Provider      string
	ProviderID    string
}

// CreateUserParams contains user information needed for account creation.
type CreateUserParams struct {
	Email     string
	Name      string
	Service   string
	AvatarURL string
}
