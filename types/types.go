package types

type OauthCreate struct {
	IsOAuth       bool
	EmailVerified bool
	Provider      string
	ProviderID    string
}
type CreateUserParams struct {
	Email     string
	Name      string
	AvatarUrl string
}
