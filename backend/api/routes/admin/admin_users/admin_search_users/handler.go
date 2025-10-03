package admin_search_users

import (
	"fmt"
	"vdm/core/locals/local_keys"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	searchUsersForAdmin(c *fiber.Ctx) error
}

type handler struct {
	repo Repository
}

func (h *handler) searchUsersForAdmin(c *fiber.Ctx) error {
	query := c.Query(local_keys.UserTag)
	if len(query) < 2 {
		return &fiber.Error{Code: fiber.StatusBadRequest, Message: fmt.Sprintf("query param <%s> is required", local_keys.UserTag)}
	}

	users, err := h.repo.searchUsersByTag(query)
	if err != nil {
		return fmt.Errorf("failed to search users by tag: %v", err)
	}

	resDTO := make([]string, len(users))
	for i := range users {
		resDTO[i] = users[i].Tag
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"results": resDTO,
	})
}
