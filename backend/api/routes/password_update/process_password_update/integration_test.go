package process_password_update

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
	"vdm/core/hmac_utils"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/bcrypt"
)

type testData struct {
	user         *models.User
	cfg          env.SecurityConfig
	validToken   uuid.UUID
	expiredToken uuid.UUID
}

func loadTestData(c context.Context, t *testing.T) (container testcontainers.Container, connector database.GormConnector, data testData) {
	container, connector = test_utils.NewTestContainerConnector(c, t)
	db := connector.GormDB()

	var err error

	defer func(c context.Context, container testcontainers.Container, connector database.GormConnector) {
		if err != nil {
			cleanUpTestData(c, t, container, connector)
			t.Fatal(err)
		}
	}(c, container, connector)

	data.cfg = env.SecurityConfig{
		PasswordTokenSecret: []byte("password"),
		PasswordTokenTTL:    1 * time.Minute,
	}

	data.user = &models.User{
		Email: "pwd_update_user@example.com",
	}

	if err = db.Create(data.user).Error; err != nil {
		return
	}

	data.validToken = uuid.New()
	validUsrTok := models.UserToken{UserID: data.user.ID, Hash: hmac_utils.HashUUID(data.validToken, data.cfg.PasswordTokenSecret),
		Expiry: time.Now().Add(1 * time.Minute), Category: models.UserTokenCategoryPassword}

	data.expiredToken = uuid.New()
	expiredUsrTok := models.UserToken{UserID: data.user.ID, Hash: hmac_utils.HashUUID(data.expiredToken, data.cfg.PasswordTokenSecret),
		Expiry: time.Now().Add(-1 * time.Minute), Category: models.UserTokenCategoryPassword}

	err = db.Create([]models.UserToken{validUsrTok, expiredUsrTok}).Error

	return
}

func cleanUpTestData(c context.Context, t *testing.T, container testcontainers.Container, connector database.GormConnector) {
	if err := connector.Close(); err != nil {
		t.Logf("failed to close connector: %v", err)
	}
	if err := container.Terminate(c); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { cleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()

	Route(connector.GormDB(), data.cfg).Register(app)

	newPwd := "NewPass123!"
	body, _ := json.Marshal(RequestDTO{Token: data.validToken, NewPassword: newPwd})
	req := httptest.NewRequest(Method, Path, bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)

	// Verify DB password updated and matches new password
	var updated models.User
	if err := connector.GormDB().First(&updated, "id = ?", data.user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte(newPwd)); err != nil {
		t.Fatalf("expected password to be updated and match new password, compare error: %v", err)
	}

	var tokenCount int
	if err := connector.GormDB().Model(&models.UserToken{}).
		Where("user_id = ?", data.user.ID).
		Select("count(*)").
		Find(&tokenCount).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, tokenCount)
}

func TestIntegration_ErrGone(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { cleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()

	Route(connector.GormDB(), data.cfg).Register(app)

	newPwd := "NewPass123!"
	body, _ := json.Marshal(RequestDTO{Token: data.expiredToken, NewPassword: newPwd})
	req := httptest.NewRequest(Method, Path, bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusGone, res.StatusCode)
}
