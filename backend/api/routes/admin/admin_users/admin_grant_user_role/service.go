package admin_grant_user_role

import (
	"fmt"
	"vdm/core/models"

	"github.com/gofiber/fiber/v2"
)

type Service interface {
	grantUserRole(userTag string, roleName models.RoleName) error
}

type service struct {
	repo Repository
}

func (s *service) grantUserRole(userTag string, roleName models.RoleName) error {
	user, err := s.repo.findUserByTagWithRole(userTag, roleName)
	if err != nil {
		return fmt.Errorf("error finding user by tag: %s", err)
	}
	if user == nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("user with tag %s not found", userTag)}
	}
	if user.HasRole(roleName) {
		return &fiber.Error{Code: fiber.StatusConflict, Message: fmt.Sprintf("user with tag %s already has role %s", userTag, roleName)}
	}

	return s.repo.createUserRole(user.ID, roleName)
}
