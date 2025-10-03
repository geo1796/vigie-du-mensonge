package inquire_sign_up

import (
	"vdm/core/dependencies/env"
	"vdm/core/dependencies/mailer"
	"vdm/core/fiberx"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	Path   = "/sign-up/inquire"
	Method = fiber.MethodPost
)

func Route(cfg env.SecurityConfig, clientURL string, db *gorm.DB, mailer mailer.Mailer) *fiberx.Route {
	repo := &repository{db}
	svc := &service{
		repo:             repo,
		mailer:           mailer,
		clientURL:        clientURL,
		emailTokenSecret: cfg.EmailTokenSecret,
		emailTokenTTL:    cfg.EmailTokenTTL,
	}
	handler := &handler{svc}

	return fiberx.NewRoute(Method, Path, handler.inquireSignUp)
}
