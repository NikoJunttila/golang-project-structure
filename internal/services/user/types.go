package userService

import "errors"

type AuthServiceEnum string

const (
	Email   AuthServiceEnum = "email"
	Google  AuthServiceEnum = "google"
	Discord AuthServiceEnum = "discord"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrParamsMismatch = errors.New("necessary params not provided")
var ErrIncorrectAuthType = errors.New("user already created using a different login method (google, discord etc...)")
var ErrUserNotFound = errors.New("failed to find user")
var ErrWrongPassword = errors.New("wrong password")
var ErrTooWeakPassword = errors.New("Weak password")

func GetServiceEnumName(service AuthServiceEnum) AuthServiceEnum {
	return service
}

type OauthCreate struct {
	IsOAuth       bool
	EmailVerified bool
	Provider      string
	ProviderID    string
}
type CreateUserParams struct {
	Email     string
	Name      string
	Service   string
	AvatarUrl string
}
