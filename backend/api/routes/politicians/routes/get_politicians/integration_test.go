package get_politicians

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"slices"
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
	politicians []models.Politician
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

	data.politicians = []models.Politician{
		{FirstName: "Emmanuel", LastName: "Macron"},
		{FirstName: "Gabriel", LastName: "Attal"},
		{FirstName: "Bruno", LastName: "Retailleau"},
	}

	err = connector.GormDB().Create(&data.politicians).Error

	return
}

func TestIntegration_Success(t *testing.T) {
	ctx := context.Background()
	container, connector, data := loadTestData(ctx, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(ctx, t, container, connector) })

	app := fiberx.NewApp()
	Group(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, Path, nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var respDTO ResponseDTO
	if err = json.Unmarshal(respBody, &respDTO); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(data.politicians), len(respDTO))

	for _, politician := range data.politicians {
		assert.True(t, slices.ContainsFunc(respDTO, func(dto PoliticianDTO) bool {
			return dto.FullName == politician.FirstName+" "+politician.LastName
		}))
	}
}
