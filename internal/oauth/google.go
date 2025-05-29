package oath

import (
	"github.com/nikojunttila/community/internal/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleLoginConfig oauth2.Config
}

var redirect_URL string = "http://localhost:3000/public/google_callback"
var AppConfig Config

func GoogleConfig() oauth2.Config {
	AppConfig.GoogleLoginConfig = oauth2.Config{
		RedirectURL:  redirect_URL, //THIS NEEDS TO BE SAME AS OATH SETTINGS IN GOOGLE CLOUD
		ClientID:     util.GetEnv("OAUTH_GOOGLE_CLIENT"),
		ClientSecret: util.GetEnv("OAUTH_GOOGLE_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}

	return AppConfig.GoogleLoginConfig
}
