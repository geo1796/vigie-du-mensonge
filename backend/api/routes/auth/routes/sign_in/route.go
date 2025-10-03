package sign_in

import (
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"

	"github.com/gofiber/fiber/v2"
)

const (
	Path   = "/sign-in"
	Method = fiber.MethodPost
)

func Route(pgxProvider database.PgxProvider, cfg env.SecurityConfig) *fiberx.Route {
	repo := &pgxRepo{pool: pgxProvider.Pool(), queries: pgxProvider.Queries()}
	svc := &service{
		repo:               repo,
		accessTokenSecret:  cfg.AccessTokenSecret,
		accessTokenTTL:     cfg.AccessTokenTTL,
		refreshTokenTTL:    cfg.RefreshTokenTTL,
		refreshTokenSecret: cfg.RefreshTokenSecret,
	}
	handler := &handler{svc}

	return fiberx.NewRoute(Method, Path, handler.signIn)
}
