package admin_find_user

import (
	"fmt"
	"vdm/core/locals/local_keys"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	findUserForAdmin(c *fiber.Ctx) error
}

type handler struct {
	repo Repository
}

func (h *handler) findUserForAdmin(c *fiber.Ctx) error {
	userTag := c.Params(local_keys.UserTag)
	if len(userTag) < 6 {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: "invalid user tag"}
	}

	user, err := h.repo.findUserByTag(userTag)
	if err != nil {
		return err
	}
	if user == nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("user with tag %s not found", userTag)}
	}

	return c.Status(fiber.StatusOK).JSON(ResponseDTO{
		Tag:       user.Tag,
		CreatedAt: user.CreatedAt,
		Roles:     user.RoleNames(),
	})
}
