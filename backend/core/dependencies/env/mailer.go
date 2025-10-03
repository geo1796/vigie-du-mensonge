package env

import (
	"fmt"
	"strconv"
)

type MailerConfig struct {
	Host     string
	Port     int
	Address  string
	Password string
}

func loadMailerConfig() (MailerConfig, error) {
	mailerPort, err := strconv.Atoi(getEnv("MAILER_PORT", "587"))
	if err != nil {
		return MailerConfig{}, fmt.Errorf("failed to parse MAILER_PORT: %v", err)
	}

	return MailerConfig{
		Host:     getEnv("MAILER_HOST", ""),
		Port:     mailerPort,
		Address:  getEnv("MAILER_ADDRESS", ""),
		Password: getEnv("MAILER_PASSWORD", ""),
	}, nil
}
