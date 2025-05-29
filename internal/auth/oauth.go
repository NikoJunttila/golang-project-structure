package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/nikojunttila/community/internal/util"
)

const (
	maxAge = 86400 * 30
)

func NewAuth() {
	ClientID := util.GetEnv("OAUTH_GOOGLE_CLIENT")
	ClientSecret := util.GetEnv("OAUTH_GOOGLE_SECRET")
	redirect_URL := util.GetEnv("GOOGLE_REDIRECT")
	key := util.GetEnv("OAUTH_KEY")
	var isProd bool
	if prod := util.GetEnv("PROD"); prod == "true" {
		isProd = true
	}

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd
	store.Options.Domain = "localhost"
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store
	goth.UseProviders(google.New(ClientID, ClientSecret, redirect_URL))
}
