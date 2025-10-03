package api

import (
	"strings"
	"time"
	"vdm/api/middlewares/authorize_authed_user"
	"vdm/api/middlewares/locals_authed_user"
	"vdm/api/routes/admin"
	"vdm/api/routes/articles"
	"vdm/api/routes/auth"
	"vdm/api/routes/get_csrf"
	"vdm/api/routes/moderator"
	"vdm/api/routes/password_update"
	"vdm/api/routes/politicians"
	"vdm/api/routes/redactor"
	"vdm/core/dependencies"
	"vdm/core/fiberx"
	"vdm/core/locals/local_keys"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/utils"
)

const Prefix = "/api/v1"

func Group(deps *dependencies.Dependencies) *fiberx.Group {
	group := fiberx.NewGroup(Prefix)

	group.Add(
		fiberx.NewMiddleware(helmet.New(helmet.Config{
			XFrameOptions:  "DENY", // stricter than SAMEORIGIN
			ReferrerPolicy: "strict-origin-when-cross-origin",
			// Minimal, valid CSP for API responses (avoid literal "none")
			ContentSecurityPolicy:     "default-src 'none'; frame-ancestors 'none'; base-uri 'none'",
			CrossOriginResourcePolicy: "same-site",
			CrossOriginOpenerPolicy:   "same-origin",
			// Disable COEP for the API to avoid breaking cross-origin fetch/embeds
			CrossOriginEmbedderPolicy: "unsafe-none",
			PermissionPolicy:          "geolocation=(), camera=(), microphone=(), payment=(), usb=()",
		})),
		fiberx.NewMiddleware(csrf.New(csrf.Config{
			KeyLookup:      "header:X-Csrf-Token",
			CookieName:     deps.Config.Security.CsrfCookieName,
			CookieSameSite: deps.Config.Security.CookieSameSite,
			CookieSecure:   deps.Config.Security.CookieSecure,
			CookieHTTPOnly: true,
			SingleUseToken: true,
			CookiePath:     "/",
			Expiration:     1 * time.Minute,
			KeyGenerator:   utils.UUIDv4,
			ContextKey:     local_keys.CsrfToken,
		})),
	)

	if deps.Config.ActiveProfile != "prod" {
		group.Add(
			fiberx.NewMiddleware(cors.New(cors.Config{
				AllowOrigins: deps.Config.ClientURL,
				AllowHeaders: strings.Join([]string{
					fiber.HeaderAccessControlAllowOrigin,
					fiber.HeaderOrigin,
					fiber.HeaderContentType,
					fiber.HeaderAccept,
					fiber.HeaderCookie,
					"X-Csrf-Token",
				}, ","),
				AllowCredentials: true,
			})),
		)
	}

	group.Add(
		get_csrf.Group(),

		/* /get-csrf has its own rate limiter */
		fiberx.NewMiddleware(limiter.New(limiter.Config{
			Max:        60, //TODO set to 30 in prod
			Expiration: 1 * time.Minute,
		})),

		auth.Group(deps),
		password_update.Group(deps),
		politicians.Group(deps),
		articles.Group(deps),

		locals_authed_user.Middleware(deps.Config.Security),
		authorize_authed_user.Middleware(deps.GormDB()),

		redactor.Group(deps),
		moderator.Group(deps),
		admin.Group(deps),
	)

	return group
}

func Register(router fiber.Router, deps *dependencies.Dependencies) {
	Group(deps).Register(router)
}
