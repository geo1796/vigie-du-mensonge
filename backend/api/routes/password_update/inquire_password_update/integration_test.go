package inquire_password_update

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
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	user *models.User
	cfg  env.SecurityConfig
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

	data.user = &models.User{
		ID:    uuid.New(),
		Email: "user@example.com",
	}

	err = db.Create(data.user).Error

	data.cfg = env.SecurityConfig{
		PasswordTokenTTL:    1 * time.Minute,
		PasswordTokenSecret: []byte("password"),
	}

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

	cfg := env.SecurityConfig{
		PasswordTokenSecret: []byte("secret"),
		PasswordTokenTTL:    1 * time.Minute,
	}
	clientURL := "https://client.example"

	mockCtrl := gomock.NewController(t)
	mailer := test_utils.NewMockMailer(mockCtrl)

	Route(cfg, clientURL, connector.GormDB(), mailer).Register(app)

	body, _ := json.Marshal(RequestDTO{Email: data.user.Email})
	req := httptest.NewRequest(Method, Path, bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	mailer.EXPECT().Send(data.user.Email, gomock.Any(), gomock.Any()).Return(nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
}

func TestIntegration_UserNotFound_ReturnsSuccess(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { cleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()

	clientURL := "https://client.example"

	mockCtrl := gomock.NewController(t)
	mailer := test_utils.NewMockMailer(mockCtrl)

	Route(data.cfg, clientURL, connector.GormDB(), mailer).Register(app)

	body, _ := json.Marshal(RequestDTO{Email: "notfound@user.com"})
	req := httptest.NewRequest(Method, Path, bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
}
