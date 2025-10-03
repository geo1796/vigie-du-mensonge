package admin_search_users

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
	users []*models.User
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

	// roles are not required for search results; create some users with tags
	data.users = []*models.User{
		{Email: "alice@test.com", Tag: "alice01", Password: "x"},
		{Email: "alicia@test.com", Tag: "alicia02", Password: "x"},
		{Email: "bob@test.com", Tag: "bob03", Password: "x"},
	}
	for _, u := range data.users {
		if err = connector.GormDB().Create(u).Error; err != nil {
			return
		}
	}
	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	// search for tags containing "ali"
	req := httptest.NewRequest(Method, Path+"?userTag=ali", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	var body struct {
		Results []string `json:"results"`
	}
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&body); err != nil {
		t.Fatal(err)
	}

	// should contain alice01 and alicia02 but not bob03
	assert.Contains(t, body.Results, data.users[0].Tag)
	assert.Contains(t, body.Results, data.users[1].Tag)
	assert.NotContains(t, body.Results, data.users[2].Tag)
}

func TestIntegration_ErrBadRequest(t *testing.T) {
	c := context.Background()
	container, connector, _ := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	// query too short (<2)
	req := httptest.NewRequest(Method, Path+"?userTag=a", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)
}
