package inquire_sign_up

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
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	oldUser *models.User
	newUser *models.User
	cfg     env.SecurityConfig
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
		EmailTokenSecret: []byte("email"),
		EmailTokenTTL:    1 * time.Minute,
	}

	data.oldUser = &models.User{Email: "old@user.com"}
	data.newUser = &models.User{Email: "new@user.com"}

	err = db.Create(data.oldUser).Error

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

func TestIntegration_Success_NewUser(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { cleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()

	mockCtrl := gomock.NewController(t)
	mailer := test_utils.NewMockMailer(mockCtrl)

	Route(data.cfg, "", connector.GormDB(), mailer).Register(app)

	reqBody, err := json.Marshal(RequestDTO{Email: data.newUser.Email})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(Method, Path, bytes.NewReader(reqBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	mailer.EXPECT().Send(data.newUser.Email, "Inscription sur Vigie du mensonge", gomock.Any()).Return(nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
}

func TestIntegration_Success_OldUser(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { cleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()

	mockCtrl := gomock.NewController(t)
	mailer := test_utils.NewMockMailer(mockCtrl)

	Route(data.cfg, "", connector.GormDB(), mailer).Register(app)

	reqBody, err := json.Marshal(RequestDTO{Email: data.oldUser.Email})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(Method, Path, bytes.NewReader(reqBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	mailer.EXPECT().Send(data.oldUser.Email, "Votre compte Vigie du mensonge", gomock.Any()).Return(nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
}
