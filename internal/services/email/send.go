package email

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v5"
	"github.com/nikojunttila/community/internal/util"
	"github.com/rs/zerolog/log"
)

type Emailer struct {
	mg     *mailgun.Client
	ApiKey string
	Domain string
}

var Mailer Emailer

func EmailerInit(cfg *Emailer) {
	cfg.Domain = util.GetEnv("MAILGUN_DOMAIN")
	cfg.ApiKey = util.GetEnv("MAILGUN_APIKEY")
	//create instance of mailgun client
	cfg.mg = mailgun.NewMailgun(cfg.ApiKey)
}

// Your available domain names can be found here:
// (https://app.mailgun.com/app/domains)

// You can find the Private API Key in your Account Menu, under "Settings":
// (https://app.mailgun.com/settings/api_security)
const SENDER = "Mailgun Sandbox <postmaster@sandbox7d11108326a74cf69ccfa984fc064eef.mailgun.org>"

// Send transmits an email to a single recipient.
// It supports both HTML and plain-text content. Providing both is highly recommended
// for maximum compatibility with different email clients.
//
// The context can be used to set a timeout or deadline for the API call. If the
// provided context does not have a deadline, a default 10-second timeout is applied.
//
// On success, it returns the message ID from Mailgun. On failure, it returns an error.
// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) to limit context to 5 seconds before ending send or other limits for cancelling email send wtih context
func (m *Emailer) Send(ctx context.Context, sender, recipient, subject, htmlContent, textContent string) (string, error) {
	// --- Input Validation ---
	if recipient == "" {
		return "", errors.New("recipient must be provided")
	}
	if subject == "" {
		return "", errors.New("subject must be provided")
	}
	if htmlContent == "" && textContent == "" {
		return "", errors.New("either htmlContent or textContent must be provided for the email body")
	}
	if sender == "" {
		sender = SENDER
	}

	// --- Message Creation ---
	// The message object allows you to add attachments, CC, BCC, and more if needed.
	// We initialize it with the required fields.
	message := mailgun.NewMessage(m.Domain, sender, subject, textContent, recipient)
	// Set the HTML part of the message if it's available.
	if htmlContent != "" {
		message.SetHTML(htmlContent)
	}
	// --- Context and Timeout Management ---
	// It's a good practice to ensure network calls have a timeout.
	// If the caller hasn't provided a context with a deadline, we'll add a default one.
	// This prevents the request from hanging indefinitely.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	// --- Sending the Email ---
	// Send the message with the specified timeout.
	resp, err := m.mg.Send(ctx, message)
	if err != nil {
		// Log the underlying error for debugging but return a more general error
		// to the caller. This prevents leaking implementation details.
		log.Error().Err(err).Str("recipient", recipient).Str("domain", sender).Msg("Failed to send email via Mailgun")
		return "", fmt.Errorf("mailgun send failed: %w", err)
	}

	// Log the successful delivery for auditing and debugging purposes.
	log.Info().Str("recipient", recipient).Str("id", resp.ID).Str("response", resp.Message).Msg("Email sent successfully")

	return resp.ID, nil
}
