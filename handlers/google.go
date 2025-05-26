package handlers

import (
	"fmt"
	"io"
	"net/http"

	// "github.com/nikojunttila/community/db"
	"github.com/nikojunttila/community/services"
)

// Redirects user to Google's OAuth2 login page
func GetGoogleLogin(w http.ResponseWriter, r *http.Request) {
	cfg := services.GoogleConfig()
	url := cfg.AuthCodeURL("randstate") // In production, generate a real state and store in cookie or session

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GetGoogleCallBack(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state != "randstate" {
		http.Error(w, "States don't Match!!", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")

	cfg := services.GoogleConfig()

	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Code-Token Exchange Failed", http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "User Data Fetch Failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "JSON Parsing Failed", http.StatusInternalServerError)
		return
	}
	fmt.Println(string(userData))
	RespondWithJson(w, 200, userData)
}
