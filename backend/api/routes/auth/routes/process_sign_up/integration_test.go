package process_sign_up

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
	"vdm/core/jwt_utils"
	"vdm/core/locals"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

var testUser = &models.User{Email: "user0@email.com"}

func loadTestData(c context.Context, t *testing.T) (container testcontainers.Container, connector database.GormConnector) {
	container, connector = test_utils.NewTestContainerConnector(c, t)

	db := connector.GormDB()

	var err error

	defer func(c context.Context, container testcontainers.Container, connector database.GormConnector) {
		if err != nil {
			cleanupTestData(c, t, container, connector)
			t.Fatal(err)
		}
	}(c, container, connector)

	if err = db.Create(&models.Role{Name: models.RoleRedactor}).Error; err != nil {
		return
	}

	err = db.Create(testUser).Error

	return
}

func cleanupTestData(c context.Context, t *testing.T, container testcontainers.Container, connector database.GormConnector) {
	if err := connector.Close(); err != nil {
		t.Logf("failed to close connector: %v", err)
	}

	if err := container.Terminate(c); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector := loadTestData(c, t)
	t.Cleanup(func() { cleanupTestData(c, t, container, connector) })

	app := fiberx.NewApp()

	dummyCfg := env.SecurityConfig{AccessTokenSecret: []byte("access"), AccessTokenTTL: 1 * time.Minute, RefreshTokenTTL: 1 * time.Minute,
		EmailTokenSecret: []byte("email"), EmailTokenTTL: 1 * time.Minute}

	Route(connector.GormDB(), dummyCfg).Register(app)

	email := "user1@email.com"
	reqDTO := RequestDTO{Username: "hello", Password: "Test123!"}

	if jwt, err := jwt_utils.GenerateJWT(
		locals.AuthedUser{ID: uuid.Nil, Email: email},
		dummyCfg.EmailTokenSecret,
		time.Now().Add(dummyCfg.EmailTokenTTL),
	); err != nil {
		t.Fatal(err)
	} else {
		reqDTO.Token = jwt
	}

	reqBody, err := json.Marshal(reqDTO)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(Method, Path, bytes.NewReader(reqBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusCreated, res.StatusCode)

	var tokenCount int64
	if err = connector.GormDB().Model(&models.UserToken{}).
		Where("user_id <> ? AND category = ?", testUser.ID, models.UserTokenCategoryRefresh).
		Count(&tokenCount).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), tokenCount)
}
