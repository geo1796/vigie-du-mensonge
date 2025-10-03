package set_auth_cookies

import (
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
)

func Middleware(cfg env.SecurityConfig) *fiberx.Middleware {
	handler := &handler{
		cookieSecure:      cfg.CookieSecure,
		cookieSameSite:    cfg.CookieSameSite,
		accessCookieName:  cfg.AccessCookieName,
		refreshCookieName: cfg.RefreshCookieName,
	}

	return fiberx.NewMiddleware(handler.setAuthCookies)
}
