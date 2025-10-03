package admin_revoke_user_role

import (
	"errors"
	"fmt"
	"vdm/core/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	findUserByTagWithRoles(userTag string) (*models.User, error)
	deleteUserRole(userID uuid.UUID, roleName models.RoleName) error
}

type repository struct {
	db *gorm.DB
}

func (r *repository) findUserByTagWithRoles(userTag string) (*models.User, error) {
	var user models.User

	if err := r.db.Where("tag = ?", userTag).
		Preload("Roles").
		Select("id").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *repository) deleteUserRole(userID uuid.UUID, roleName models.RoleName) error {
	var role models.Role
	if err := r.db.Where("name = ?", roleName).
		Select("id").
		First(&role).Error; err != nil {
		return fmt.Errorf("error finding Role{name=%s}: %v", roleName, err)
	}

	if err := r.db.Delete(&models.UserRole{UserID: userID, RoleID: role.ID}).Error; err != nil {
		return fmt.Errorf("error deleting UserRole{UserID=%s, RoleID=%s}: %v", userID, role.ID, err)
	}

	return nil
}
