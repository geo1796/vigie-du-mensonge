package locals_authed_user

import (
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
)

func Middleware(cfg env.SecurityConfig) *fiberx.Middleware {
	handler := &handler{
		accessTokenSecret: cfg.AccessTokenSecret,
		accessCookieName:  cfg.AccessCookieName,
	}

	return fiberx.NewMiddleware(handler.localsAuthedUser)
}
