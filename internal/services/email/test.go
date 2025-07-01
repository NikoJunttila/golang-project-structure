package email

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v5"
	"github.com/rs/zerolog/log"
)

// Your available domain names can be found here:
// (https://app.mailgun.com/app/domains)

// You can find the Private API Key in your Account Menu, under "Settings":
// (https://app.mailgun.com/settings/api_security)

func (m *Emailer)SendTestEmail() {
	// When you have an EU domain, you must specify the endpoint:
	// err := mg.SetAPIBase(mailgun.APIBaseEU)
	sender := SENDER
	subject := "Fancy subject with stuff from env!"
	body := "Test message"
	recipient := "nikosamulijunttila@gmail.com"

	// The message object allows you to add attachments and Bcc recipients
	message := mailgun.NewMessage(m.Domain, sender, subject, body, recipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10-second timeout
	resp, err := m.mg.Send(ctx, message)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Info().Str("email_id", resp.ID).Msgf("sent email %s", resp.Message)
}
