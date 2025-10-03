package sign_in

import (
	"fmt"
	"time"
	"vdm/core/hmac_utils"
	"vdm/core/jwt_utils"
	"vdm/core/locals"
	"vdm/core/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	signIn(req RequestDTO) (models.User, locals.AccessToken, locals.RefreshToken, error)
}

type service struct {
	accessTokenSecret []byte
	accessTokenTTL    time.Duration

	refreshTokenSecret []byte
	refreshTokenTTL    time.Duration

	repo Repository
}

func (s *service) signIn(req RequestDTO) (models.User, locals.AccessToken, locals.RefreshToken, error) {
	user, err := s.repo.findUserByEmail(req.Email)
	if err != nil {
		return models.User{}, locals.AccessToken{}, locals.RefreshToken{}, &fiber.Error{Code: fiber.StatusUnauthorized, Message: "user not found"}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return models.User{}, locals.AccessToken{}, locals.RefreshToken{}, &fiber.Error{Code: fiber.StatusUnauthorized, Message: "wrong password"}
	}

	rft := uuid.New()

	usrTok := models.UserToken{
		UserID:   user.ID,
		Expiry:   time.Now().Add(s.refreshTokenTTL),
		Hash:     hmac_utils.HashUUID(rft, s.refreshTokenSecret),
		Category: models.UserTokenCategoryRefresh,
	}

	if err = s.repo.createRefreshToken(&usrTok); err != nil {
		return models.User{}, locals.AccessToken{}, locals.RefreshToken{}, fmt.Errorf("failed to create refresh rft: %v", err)
	}

	jwtExpiry := time.Now().Add(s.accessTokenTTL)
	jwt, err := jwt_utils.GenerateJWT(locals.AuthedUser{ID: user.ID, Email: user.Email},
		s.accessTokenSecret, jwtExpiry)
	if err != nil {
		return models.User{}, locals.AccessToken{}, locals.RefreshToken{}, fmt.Errorf("failed to generate JWT: %v", err)
	}

	return user,
		locals.AccessToken{Token: jwt, Expiry: jwtExpiry},
		locals.RefreshToken{Token: rft, Expiry: usrTok.Expiry},
		nil
}
