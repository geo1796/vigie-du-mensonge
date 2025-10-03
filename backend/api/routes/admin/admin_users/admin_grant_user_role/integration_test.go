package admin_grant_user_role

import (
	"context"
	"net/http/httptest"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/fiberx"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	roles         map[models.RoleName]models.Role
	userNoRole    *models.User
	userModerator *models.User
	userAdmin     *models.User
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

	// create roles
	roles := []models.Role{{Name: models.RoleAdmin}, {Name: models.RoleModerator}, {Name: models.RoleRedactor}}
	if err = connector.GormDB().Create(&roles).Error; err != nil {
		return
	}
	data.roles = make(map[models.RoleName]models.Role)
	for i := range roles {
		data.roles[roles[i].Name] = roles[i]
	}

	// users
	data.userNoRole = &models.User{Email: "no_role@test.com", Tag: "user_no_role", Password: "x"}
	if err = connector.GormDB().Create(data.userNoRole).Error; err != nil {
		return
	}

	data.userModerator = &models.User{Email: "has_moderator@test.com", Tag: "user_has_mod", Password: "x", Roles: []*models.Role{&roles[1]}}
	if err = connector.GormDB().Create(data.userModerator).Error; err != nil {
		return
	}

	data.userAdmin = &models.User{Email: "admin@test.com", Tag: "admin_user", Password: "x", Roles: []*models.Role{&roles[0]}}
	err = connector.GormDB().Create(data.userAdmin).Error
	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	// grant MODERATOR to user without roles
	req := httptest.NewRequest(Method, "/"+data.userNoRole.Tag+"/roles/"+string(models.RoleModerator), nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}

	var cnt int64
	if err := connector.GormDB().Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", data.userNoRole.ID, data.roles[models.RoleModerator].ID).
		Count(&cnt).Error; err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Fatalf("expected user role to be created")
	}
}

func TestIntegration_ErrNotFound(t *testing.T) {
	c := context.Background()
	container, connector, _ := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/unknown_user_tag/roles/"+string(models.RoleModerator), nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestIntegration_ErrConflict(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	// user already has MODERATOR
	req := httptest.NewRequest(Method, "/"+data.userModerator.Tag+"/roles/"+string(models.RoleModerator), nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", res.StatusCode)
	}
}

func TestIntegration_ErrBadRequest(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/"+data.userNoRole.Tag+"/roles/"+string(models.RoleAdmin), nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
