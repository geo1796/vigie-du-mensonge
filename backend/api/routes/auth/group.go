package auth

import (
	"vdm/api/routes/auth/middlewares/set_auth_cookies"
	"vdm/api/routes/auth/routes/inquire_sign_up"
	"vdm/api/routes/auth/routes/process_sign_up"
	"vdm/api/routes/auth/routes/refresh"
	"vdm/api/routes/auth/routes/sign_in"
	"vdm/api/routes/auth/routes/sign_out"
	"vdm/core/dependencies"
	"vdm/core/fiberx"
)

const Prefix = "/auth"

func Group(deps *dependencies.Dependencies) *fiberx.Group {
	group := fiberx.NewGroup(Prefix)

	group.Add(
		inquire_sign_up.Route(deps.Config.Security, deps.Config.ClientURL, deps.GormDB(), deps.Mailer),

		set_auth_cookies.Middleware(deps.Config.Security),

		process_sign_up.Route(deps.GormDB(), deps.Config.Security),
		sign_in.Route(deps.PgxProvider, deps.Config.Security),
		refresh.Route(deps.GormDB(), deps.Config.Security),
		sign_out.Route(deps.GormDB(), deps.Config.Security),
	)

	return group
}
