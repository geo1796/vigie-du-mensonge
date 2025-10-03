package admin_find_user

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/fiberx"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	roles map[models.RoleName]models.Role
	user  *models.User
}

func loadTestData(c context.Context, t *testing.T) (container testcontainers.Container, connector database.Connector, data testData) {
	container, connector = test_utils.NewTestContainerConnector(c, t)
	var err error
	defer func() {
		if err != nil {
			test_utils.CleanUpTestData(c, t, container, connector)
			t.Fatal(err)
		}
	}()

	// roles
	roles := []models.Role{{Name: models.RoleAdmin}, {Name: models.RoleModerator}, {Name: models.RoleRedactor}}
	if err = connector.GormDB().Create(&roles).Error; err != nil {
		return
	}
	data.roles = make(map[models.RoleName]models.Role)
	for i := range roles {
		data.roles[roles[i].Name] = roles[i]
	}

	// user with two roles
	data.user = &models.User{Email: "user@test.com", Tag: "user_find_me", Password: "x", Roles: []*models.Role{&roles[1], &roles[2]}}
	err = connector.GormDB().Create(data.user).Error
	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/"+data.user.Tag, nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	var dto ResponseDTO
	if err := json.NewDecoder(res.Body).Decode(&dto); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, data.user.Tag, dto.Tag)
	assert.NotZero(t, dto.CreatedAt)
	// roles order is not guaranteed; assert contents
	assert.ElementsMatch(t, []models.RoleName{models.RoleModerator, models.RoleRedactor}, dto.Roles)
}

func TestIntegration_ErrNotFound(t *testing.T) {
	c := context.Background()
	container, connector, _ := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/unknown_tag", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, res.StatusCode)
}

func TestIntegration_ErrBadRequest(t *testing.T) {
	c := context.Background()
	container, connector, _ := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	// missing path param -> use root path of this route group
	req := httptest.NewRequest(Method, "/abc", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)
}
