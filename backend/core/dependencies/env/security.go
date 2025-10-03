package env

import (
	"fmt"
	"time"
)

type SecurityConfig struct {
	AccessTokenSecret []byte
	AccessTokenTTL    time.Duration

	RefreshTokenSecret []byte
	RefreshTokenTTL    time.Duration

	EmailTokenSecret []byte
	EmailTokenTTL    time.Duration

	PasswordTokenSecret []byte
	PasswordTokenTTL    time.Duration

	RefreshCookieName string
	AccessCookieName  string
	CsrfCookieName    string
	CookieSameSite    string
	CookieSecure      bool
}

func loadSecurityConfig() (SecurityConfig, error) {
	accessTokenTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "45s"))
	if err != nil {
		return SecurityConfig{}, fmt.Errorf("failed to parse ACCESS_TOKEN_TTL: %v", err)
	}

	refreshTokenTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "60s"))
	if err != nil {
		return SecurityConfig{}, fmt.Errorf("failed to parse REFRESH_TOKEN_TTL: %v", err)
	}

	emailTokenTTL, err := time.ParseDuration(getEnv("EMAIL_TOKEN_TTL", "45s"))
	if err != nil {
		return SecurityConfig{}, fmt.Errorf("failed to parse EMAIL_TOKEN_TTL: %v", err)
	}

	passwordTokenTTL, err := time.ParseDuration(getEnv("PASSWORD_TOKEN_TTL", "45s"))
	if err != nil {
		return SecurityConfig{}, fmt.Errorf("failed to parse PASSWORD_TOKEN_TTL: %v", err)
	}

	return SecurityConfig{
		EmailTokenSecret:    []byte(getEnv("EMAIL_TOKEN_SECRET", "")),
		EmailTokenTTL:       emailTokenTTL,
		PasswordTokenSecret: []byte(getEnv("PASSWORD_TOKEN_SECRET", "")),
		PasswordTokenTTL:    passwordTokenTTL,
		AccessTokenSecret:   []byte(getEnv("ACCESS_TOKEN_SECRET", "")),
		AccessTokenTTL:      accessTokenTTL,
		RefreshTokenSecret:  []byte(getEnv("REFRESH_TOKEN_SECRET", "")),
		RefreshTokenTTL:     refreshTokenTTL,
		AccessCookieName:    getEnv("ACCESS_COOKIE_NAME", "__Host-jwt"),
		RefreshCookieName:   getEnv("REFRESH_COOKIE_NAME", "__Host-rft"),
		CookieSameSite:      getEnv("COOKIE_SAME_SITE", "strict"),
		CookieSecure:        getEnv("COOKIE_SECURE", "true") == "true",
		CsrfCookieName:      getEnv("CSRF_COOKIE_NAME", "__Host-csrf_"),
	}, nil
}
