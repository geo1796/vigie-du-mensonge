package admin_revoke_user_role

import (
	"vdm/core/fiberx"
	"vdm/core/locals/local_keys"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	Path   = "/:" + local_keys.UserTag + "/roles/:" + local_keys.RoleName
	Method = fiber.MethodDelete
)

func Route(db *gorm.DB) *fiberx.Route {
	repo := &repository{db}
	svc := &service{repo}
	handler := &handler{svc}
	return fiberx.NewRoute(Method, Path, handler.revokeUserRoleForAdmin)
}
