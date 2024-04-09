package email

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridEmailProvider struct {
	ApiKey string
}

func NewSendGridEmailProvider(apiKey string) *SendGridEmailProvider {
	return &SendGridEmailProvider{ApiKey: apiKey}
}

func (s *SendGridEmailProvider) Send(fromAddress string, fromName, toAddress string, subject string, content string) error {
	from := mail.NewEmail(fromName, fromAddress)
	to := mail.NewEmail("", toAddress)
	message := mail.NewSingleEmail(from, subject, to, "", content)
	client := sendgrid.NewSendClient(s.ApiKey)
	_, err := client.Send(message)
	return err
}
