package refresh

import (
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	Path   = "/refresh"
	Method = fiber.MethodPost
)

func Route(db *gorm.DB, cfg env.SecurityConfig) *fiberx.Route {
	repo := &repository{db}

	svc := &service{
		repo:               repo,
		accessTokenTTL:     cfg.AccessTokenTTL,
		refreshTokenTTL:    cfg.RefreshTokenTTL,
		accessTokenSecret:  cfg.AccessTokenSecret,
		refreshTokenSecret: cfg.RefreshTokenSecret,
	}

	handler := &handler{
		svc:               svc,
		refreshCookieName: cfg.RefreshCookieName,
	}

	return fiberx.NewRoute(Method, Path, handler.refresh)
}
