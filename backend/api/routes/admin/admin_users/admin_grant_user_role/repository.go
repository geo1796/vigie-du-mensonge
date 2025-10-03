package admin_grant_user_role

import (
	"errors"
	"fmt"
	"vdm/core/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	findUserByTagWithRole(userTag string, roleName models.RoleName) (*models.User, error)
	createUserRole(userID uuid.UUID, roleName models.RoleName) error
}

type repository struct {
	db *gorm.DB
}

func (r *repository) findUserByTagWithRole(userTag string, roleName models.RoleName) (*models.User, error) {
	var user models.User

	if err := r.db.Where("tag = ?", userTag).
		Preload("Roles", "name = ?", roleName).
		Select("id").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *repository) createUserRole(userID uuid.UUID, roleName models.RoleName) error {
	var role models.Role
	if err := r.db.Where("name = ?", roleName).
		Select("id").
		First(&role).Error; err != nil {
		return fmt.Errorf("error finding Role{Name=%s}: %v", roleName, err)
	}

	if err := r.db.Create(&models.UserRole{UserID: userID, RoleID: role.ID}).Error; err != nil {
		return fmt.Errorf("error creating UserRole{UserID=%s RoleID=%s}: %v", userID, role.ID, err)
	}

	return nil
}
