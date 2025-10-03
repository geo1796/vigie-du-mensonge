package admin_revoke_user_role

import (
	"fmt"
	"vdm/core/models"

	"github.com/gofiber/fiber/v2"
)

type Service interface {
	revokeUserRole(userTag string, roleName models.RoleName) error
}

type service struct {
	repo Repository
}

func (s *service) revokeUserRole(userTag string, roleName models.RoleName) error {
	user, err := s.repo.findUserByTagWithRoles(userTag)
	if err != nil {
		return err
	}
	if user == nil {
		return &fiber.Error{Code: fiber.StatusNotFound, Message: fmt.Sprintf("user with tag %s not found", userTag)}
	}
	if user.HasRole(models.RoleAdmin) {
		return &fiber.Error{Code: fiber.StatusForbidden, Message: fmt.Sprintf("user with tag %s is ADMIN", userTag)}
	}
	if !user.HasRole(roleName) {
		return &fiber.Error{Code: fiber.StatusConflict, Message: fmt.Sprintf("user with tag %s has no role %s", userTag, roleName)}
	}

	return s.repo.deleteUserRole(user.ID, roleName)
}
