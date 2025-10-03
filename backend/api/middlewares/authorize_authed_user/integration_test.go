package authorize_authed_user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vdm/api/middlewares/locals_authed_user"
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
	"vdm/core/jwt_utils"
	"vdm/core/locals"
	"vdm/core/locals/local_keys"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	roles []*models.Role
	user  *models.User
	cfg   env.SecurityConfig
}

func loadTestData(c context.Context, t *testing.T) (container testcontainers.Container, connector database.GormConnector, data testData) {
	container, connector = test_utils.NewTestContainerConnector(c, t)

	var err error

	defer func() {
		if err != nil {
			test_utils.CleanUpTestData(c, t, container, connector)
			t.Fatal(err)
		}
	}()

	data.cfg = env.SecurityConfig{
		AccessTokenSecret: []byte("access"),
		AccessTokenTTL:    1 * time.Minute,
		AccessCookieName:  "jwt",
	}

	data.roles = []*models.Role{
		{Name: models.RoleAdmin},
		{Name: models.RoleModerator},
		{Name: models.RoleRedactor},
	}

	if err = connector.GormDB().Create(data.roles).Error; err != nil {
		return
	}

	data.user = &models.User{Email: "user@test.com", Roles: []*models.Role{data.roles[0]}}

	err = connector.GormDB().Create(data.user).Error
	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	locals_authed_user.Middleware(data.cfg).Register(app)
	Middleware(connector.GormDB()).Register(app)
	app.Use(func(c *fiber.Ctx) error {
		authedUser, ok := c.Locals(local_keys.AuthedUser).(locals.AuthedUser)
		if !ok {
			return fiber.ErrInternalServerError
		}

		assert.Equal(t, 1, len(authedUser.Roles))
		assert.Equal(t, models.RoleAdmin, authedUser.Roles[0])

		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)

	if jwt, err := jwt_utils.GenerateJWT(locals.AuthedUser{ID: data.user.ID},
		data.cfg.AccessTokenSecret, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	} else {
		req.AddCookie(&http.Cookie{Name: data.cfg.AccessCookieName, Value: jwt})
	}

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
}
