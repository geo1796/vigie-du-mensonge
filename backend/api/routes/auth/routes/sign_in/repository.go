package sign_in

import (
	"context"
	"vdm/core/dependencies/database/queries"
	"vdm/core/logger"
	"vdm/core/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

type Repository interface {
	findUserByEmail(email string) (models.User, error)
	createRefreshToken(rft *models.UserToken) error
}

type gormRepo struct {
	db *gorm.DB
}

func (r *gormRepo) findUserByEmail(email string) (models.User, error) {
	var user models.User
	if err := r.db.Model(&models.User{}).
		Where("email = ?", email).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Select("id", "email", "password", "tag", "created_at", "updated_at").
		First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *gormRepo) createRefreshToken(rft *models.UserToken) (err error) {
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

type pgxRepo struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func (r *pgxRepo) findUserByEmail(email string) (models.User, error) {
	ctx := context.Background()
	urow, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		ID:       urow.ID,
		Email:    urow.Email,
		Password: urow.Password,
		Tag:      urow.Tag,
	}

	roles, err := r.queries.GetRolesByUserID(ctx, urow.ID)
	if err != nil {
		return models.User{}, err
	}
	if len(roles) > 0 {
		user.Roles = make([]*models.Role, 0, len(roles))
		for _, rr := range roles {
			role := &models.Role{
				ID:   rr.ID,
				Name: models.RoleName(rr.Name),
			}
			user.Roles = append(user.Roles, role)
		}
	}
	return user, nil
}

func (r *pgxRepo) createRefreshToken(rft *models.UserToken) (err error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				logger.Error("failed to rollback transaction", logger.Err(rbErr))
			}
			return
		}
		if cmErr := tx.Commit(ctx); cmErr != nil {
			logger.Error("failed to commit transaction", logger.Err(cmErr))
			err = cmErr
		}
	}()

	qtx := r.queries.WithTx(tx)

	// delete existing refresh tokens for the user
	if err = qtx.DeleteRefreshTokensByUserID(ctx, rft.UserID); err != nil {
		return err
	}

	// insert new refresh token
	_, err = qtx.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		UserID:   rft.UserID,
		Category: string(rft.Category),
		Hash:     rft.Hash,
		Expiry:   rft.Expiry,
	})
	return err
}
