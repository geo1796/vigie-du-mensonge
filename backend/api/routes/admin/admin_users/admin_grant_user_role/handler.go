package admin_grant_user_role

import (
	"vdm/core/locals/local_keys"
	"vdm/core/models"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	grantUserRoleForAdmin(c *fiber.Ctx) error
}

type handler struct {
	svc Service
}

func (h *handler) grantUserRoleForAdmin(c *fiber.Ctx) error {
	userTag := c.Params(local_keys.UserTag)
	roleName := models.RoleName(c.Params(local_keys.RoleName))

	if len(userTag) < 6 || !roleName.IsValid() || roleName == models.RoleAdmin {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "invalid userTag or roleName"}
	}

	if err := h.svc.grantUserRole(userTag, roleName); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
