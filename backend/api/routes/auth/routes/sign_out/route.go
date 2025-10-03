package sign_out

import (
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	Path = "/sign-out"

	Method = fiber.MethodPost
)

func Route(db *gorm.DB, cfg env.SecurityConfig) *fiberx.Route {
	repo := &repository{db}
	service := &service{
		repo:              repo,
		accessTokenSecret: cfg.AccessTokenSecret,
	}
	handler := &handler{svc: service, accessCookieName: cfg.AccessCookieName}

	return fiberx.NewRoute(Method, Path, handler.signOut)
}
