package process_sign_up

import (
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	Path   = "/sign-up/process"
	Method = fiber.MethodPost
)

func Route(db *gorm.DB, cfg env.SecurityConfig) *fiberx.Route {
	repo := &repository{db}
	svc := &service{
		repo:               repo,
		accessTokenSecret:  cfg.AccessTokenSecret,
		accessTokenTTL:     cfg.AccessTokenTTL,
		refreshTokenTTL:    cfg.RefreshTokenTTL,
		refreshTokenSecret: cfg.RefreshTokenSecret,
		emailTokenSecret:   cfg.EmailTokenSecret,
	}
	handler := &handler{svc: svc}

	return fiberx.NewRoute(Method, Path, handler.processSignUp)
}
