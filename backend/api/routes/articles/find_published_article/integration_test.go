package find_published_article

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/dto/response_dto"
	"vdm/core/fiberx"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	politicians []*models.Politician
	article     *models.Article
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

	data.politicians = []*models.Politician{{FirstName: "Emmanuel", LastName: "Macron"}}
	if err = connector.GormDB().Create(&data.politicians).Error; err != nil {
		return
	}

	redactor := &models.User{Email: "redactor@test.com", Tag: "redactor0123", Password: "x"}
	if err = connector.GormDB().Create(redactor).Error; err != nil {
		return
	}

	data.article = &models.Article{
		RedactorID:  redactor.ID,
		Title:       "Article about Emmanuel Macron",
		Politicians: []*models.Politician{data.politicians[0]},
		Tags:        []*models.ArticleTag{{Tag: "Emmanuel Macron"}},
		Status:      models.ArticleStatusPublished,
		// other fields have defaults in schema; keep minimal like other tests
	}

	err = connector.GormDB().Create(data.article).Error
	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/"+data.article.ID.String(), nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("Expected status code 200, got %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var resDTO response_dto.Article
	if err = json.Unmarshal(resBody, &resDTO); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, data.article.ID, resDTO.ID)
	assert.Equal(t, 1, len(resDTO.Politicians))
	assert.Equal(t, 1, len(resDTO.Tags))
}

func TestIntegration_ErrNotFound(t *testing.T) {
	c := context.Background()
	container, connector, _ := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/"+uuid.New().String(), nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, res.StatusCode)
}

func TestIntegration_BadRequest(t *testing.T) {
	c := context.Background()
	container, connector, _ := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/not-a-uuid", nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)
}
