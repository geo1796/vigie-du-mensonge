package admin_revoke_user_role

import (
	"vdm/core/locals/local_keys"
	"vdm/core/models"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	revokeUserRoleForAdmin(c *fiber.Ctx) error
}

type handler struct {
	svc Service
}

func (h *handler) revokeUserRoleForAdmin(c *fiber.Ctx) error {
	userTag := c.Params(local_keys.UserTag)
	roleName := models.RoleName(c.Params(local_keys.RoleName))

	if userTag == "" || !roleName.IsValid() || roleName == models.RoleAdmin {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "invalid userTag or roleName"}
	}

	if err := h.svc.revokeUserRole(userTag, roleName); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
