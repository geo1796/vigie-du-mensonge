package mailer

import (
	"vdm/core/dependencies/env"
	"vdm/core/logger"

	"gopkg.in/gomail.v2"
)

type Mailer interface {
	Send(to, subject, body string) error
}

type mailer struct {
	*gomail.Dialer
	from string
}

func New(config env.MailerConfig) Mailer {
	return &mailer{
		Dialer: gomail.NewDialer(config.Host, config.Port, config.Address, config.Password),
		from:   config.Address,
	}
}

func (m *mailer) Send(to, subject, body string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	err := m.DialAndSend(msg)
	if err != nil {
		logger.Error("error sending email", logger.Err(err))
		return err
	}

	return nil
}
