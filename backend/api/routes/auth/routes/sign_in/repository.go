package sign_in

import (
	"vdm/core/logger"
	"vdm/core/models"

	"gorm.io/gorm"
)

type Repository interface {
	findUserByEmail(email string) (models.User, error)
	createRefreshToken(rft *models.UserToken) error
}

type repository struct {
	db *gorm.DB
}

func (r *repository) findUserByEmail(email string) (models.User, error) {
	var user models.User
	if err := r.db.Model(&models.User{}).
		Where("email = ?", email).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Select("id", "password", "tag", "created_at", "updated_at").
		First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *repository) createRefreshToken(rft *models.UserToken) (err error) {
	tx := r.db.Begin()

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback().Error; rbErr != nil {
				logger.Error("failed to rollback transaction", logger.Err(rbErr))
			}
			return
		}

		if cmErr := tx.Commit().Error; cmErr != nil {
			logger.Error("failed to commit transaction", logger.Err(cmErr))
			err = cmErr
		}
	}()

	if err = tx.Model(&models.UserToken{}).
		Where("user_id = ? AND category = ?", rft.UserID, models.UserTokenCategoryRefresh).
		Delete(&models.UserToken{}).Error; err != nil {
		return
	}

	err = tx.Create(rft).Error
	return
}
