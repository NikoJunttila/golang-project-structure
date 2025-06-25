package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/rs/zerolog/log"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func GetHomeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		log.Error().Msgf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

type user struct {
	Username string
	Password string
	Secret   string
}

// Simple in-memory "database" for demonstration purposes
var users = map[string]*user{
	"john": {Username: "john", Password: "password", Secret: ""},
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			log.Error().Msgf("Template error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Error().Msgf("Form parse error: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.Form.Get("username"))
	password := r.Form.Get("password")

	// Validate input
	if username == "" || password == "" {
		http.Redirect(w, r, "/two/login", http.StatusFound)
		return
	}

	user, ok := users[username]
	if !ok || user.Password != password {
		http.Redirect(w, r, "/two/login", http.StatusFound)
		return
	}

	if user.Secret == "" {
		http.Redirect(w, r, "/two/generate-otp?username="+url.QueryEscape(username), http.StatusFound)
		return
	}

	err := templates.ExecuteTemplate(w, "validate.html", struct{ Username string }{Username: username})
	if err != nil {
		log.Error().Msgf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func GetGenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		http.Redirect(w, r, "/two", http.StatusFound)
		return
	}

	user, ok := users[username]
	if !ok {
		http.Redirect(w, r, "/two", http.StatusFound)
		return
	}

	// CRITICAL FIX: Only generate the secret once and save it properly
	if user.Secret == "" {
		secret, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Go2FADemo",
			AccountName: username,
		})
		if err != nil {
			log.Error().Msgf("Failed to generate TOTP secret for %s: %v", username, err)
			http.Error(w, "Failed to generate TOTP secret.", http.StatusInternalServerError)
			return
		}

		// CRITICAL: Use the Base32-encoded secret string, not raw bytes
		// And modify the original user in the map, not a copy
		users[username].Secret = secret.Secret()

		log.Info().Msgf("Generated TOTP secret for user %s: %s", username, secret.Secret())
	}

	// Build the OTP URL
	otpURL := fmt.Sprintf("otpauth://totp/Go2FADemo:%s?secret=%s&issuer=Go2FADemo",
		url.QueryEscape(username),
		user.Secret)

	qrCodeBase64, err := generateQRCodeBase64(otpURL)
	if err != nil {
		log.Error().Msgf("Failed to generate QR code for user %s: %v", username, err)
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	data := struct {
		OTPURL     string
		Username   string
		Secret     string
		QRCodeData string
	}{
		OTPURL:     otpURL,
		Username:   username,
		Secret:     user.Secret,
		QRCodeData: qrCodeBase64,
	}

	err = templates.ExecuteTemplate(w, "qrcode.html", data)
	if err != nil {
		log.Error().Msgf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func ValidateOTPHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		username := strings.TrimSpace(r.URL.Query().Get("username"))
		if username == "" {
			http.Redirect(w, r, "/two/login", http.StatusFound)
			return
		}

		err := templates.ExecuteTemplate(w, "validate.html", struct{ Username string }{Username: username})
		if err != nil {
			log.Error().Msgf("Template error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	case "POST":
		if err := r.ParseForm(); err != nil {
			log.Error().Msgf("Form parse error: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		otpCode := strings.TrimSpace(r.FormValue("otpCode"))
		otpCode = strings.ReplaceAll(otpCode, " ", "")

		// Validate input
		if username == "" || otpCode == "" {
			http.Redirect(w, r, "/two/login", http.StatusFound)
			return
		}

		user, exists := users[username]
		if !exists || user.Secret == "" {
			log.Error().Msgf("User %s does not exist or has no secret", username)
			http.Redirect(w, r, "/two/login", http.StatusFound)
			return
		}

		// CRITICAL: Debug logging to help troubleshoot
		log.Info().Msgf("Validating TOTP for user %s with code %s and secret %s", username, otpCode, user.Secret)

		// Validate the TOTP code
		isValid := totp.Validate(otpCode, user.Secret)
		if !isValid {
			log.Warn().Msgf("Invalid TOTP code for user %s: provided=%s", username, otpCode)

			// Redirect back to validation page with error
			http.Redirect(w, r, fmt.Sprintf("/two/validate-otp?username=%s&error=invalid", url.QueryEscape(username)), http.StatusTemporaryRedirect)
			return
		}

		log.Info().Msgf("TOTP validation successful for user %s", username)

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "authenticatedUser",
			Value:    username, // Store username instead of just "true"
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,  // Security improvement
			Secure:   false, // Set to true in production with HTTPS
		})

		http.Redirect(w, r, "/two/dashboard", http.StatusSeeOther)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("authenticatedUser")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// You can now use cookie.Value as the username if needed
	err = templates.ExecuteTemplate(w, "dashboard.html", struct{ Username string }{Username: cookie.Value})
	if err != nil {
		log.Error().Msgf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
func DebugHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	user, ok := users[username]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Username: %s\n", username)
	fmt.Fprintf(w, "Secret: %s\n", user.Secret)
	fmt.Fprintf(w, "Secret Length: %d\n", len(user.Secret))

	// Generate current TOTP for comparison
	code, err := totp.GenerateCode(user.Secret, time.Now())
	if err != nil {
		fmt.Fprintf(w, "Error generating code: %v\n", err)
	} else {
		fmt.Fprintf(w, "Current TOTP: %s\n", code)
	}
}

type nopWriteCloser struct {
	io.Writer
}

func (n nopWriteCloser) Close() error {
	return nil
}

// generateQRCodeBase64 generates a QR code and returns it as a base64-encoded string
func generateQRCodeBase64(data string) (string, error) {
	// Create QR code
	qrc, err := qrcode.NewWith(data, qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow))
	if err != nil {
		return "", fmt.Errorf("could not generate QRCode: %v", err)
	}
	g := standard.NewGradient(45, []standard.ColorStop{
		{
			T:     0,
			Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			T:     0.5,
			Color: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			T:     1,
			Color: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
	}...)

	options := []standard.ImageOption{
		standard.WithQRWidth(8),
		standard.WithFgGradient(g),
		standard.WithCircleShape(),
	}
	var buf bytes.Buffer
	w := nopWriteCloser{&buf}
	writer := standard.NewWithWriter(w, options...)

	// Save QR code to buffer
	if err = qrc.Save(writer); err != nil {
		return "", fmt.Errorf("could not save QR code: %v", err)
	}
	// Encode to base64
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64String, nil
}
