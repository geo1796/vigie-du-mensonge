package sign_out

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
	"vdm/core/jwt_utils"
	"vdm/core/locals"
	"vdm/core/models"
	"vdm/test_utils"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

var testUser = &models.User{ID: uuid.New(), Email: "signout_user0@email.com"}

func loadTestData(c context.Context, t *testing.T) (container testcontainers.Container, connector database.GormConnector) {
	container, connector = test_utils.NewTestContainerConnector(c, t)

	db := connector.GormDB()

	var err error

	defer func() {
		if err != nil {
			cleanupTestData(c, t, container, connector)
			t.Fatal(err)
		}
	}()

	// create a user and one refresh token
	if err = db.Create(testUser).Error; err != nil {
		return
	}
	rft := &models.UserToken{UserID: testUser.ID, Expiry: time.Now().Add(10 * time.Minute),
		Category: models.UserTokenCategoryRefresh}
	if err = db.Create(rft).Error; err != nil {
		return
	}

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

	dummyCfg := env.SecurityConfig{AccessTokenSecret: []byte("dummySecret"),
		AccessCookieName: "access"}

	Route(connector.GormDB(), dummyCfg).Register(app)

	req := httptest.NewRequest(Method, Path, nil)

	// Generate access token for the user using the same secret as in route (defaults are OK in tests)
	if jwt, err := jwt_utils.GenerateJWT(locals.AuthedUser{ID: testUser.ID}, dummyCfg.AccessTokenSecret, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	} else {
		req.AddCookie(&http.Cookie{Name: dummyCfg.AccessCookieName, Value: jwt})
	}

	var tokenCount int64
	if err := connector.GormDB().Model(&models.UserToken{}).Where("user_id = ?", testUser.ID).Count(&tokenCount).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), tokenCount)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)

	// Ensure refresh tokens are deleted for this user
	if err := connector.GormDB().Model(&models.UserToken{}).Where("user_id = ?", testUser.ID).Count(&tokenCount).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(0), tokenCount)
}
