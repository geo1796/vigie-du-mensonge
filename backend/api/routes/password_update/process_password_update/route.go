package process_password_update

import (
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	Path   = "/process"
	Method = fiber.MethodPost
)

func Route(db *gorm.DB, cfg env.SecurityConfig) *fiberx.Route {
	repo := &repository{db}
	svc := &service{
		tokenSecret: cfg.PasswordTokenSecret,
		repo:        repo,
	}
	handler := &handler{svc}

	return fiberx.NewRoute(Method, Path, handler.processPasswordUpdate)
}
