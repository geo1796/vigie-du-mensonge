package admin_users

import (
	"vdm/api/routes/admin/admin_users/admin_find_user"
	"vdm/api/routes/admin/admin_users/admin_grant_user_role"
	"vdm/api/routes/admin/admin_users/admin_revoke_user_role"
	"vdm/api/routes/admin/admin_users/admin_search_users"
	"vdm/core/dependencies"
	"vdm/core/fiberx"
)

const Prefix = "/users"

func Group(deps *dependencies.Dependencies) *fiberx.Group {
	group := fiberx.NewGroup(Prefix)

	group.Add(
		admin_search_users.Route(deps.GormDB()),
		admin_find_user.Route(deps.GormDB()),
		admin_grant_user_role.Route(deps.GormDB()),
		admin_revoke_user_role.Route(deps.GormDB()),
	)

	return group
}
