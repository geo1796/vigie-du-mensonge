package get_published_articles

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"slices"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/dto/response_dto"
	"vdm/core/fiberx"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	politicians []*models.Politician
	articles    []*models.Article
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

	data.politicians = []*models.Politician{
		{FirstName: "Nicolas", LastName: "Sarkozy"},
		{FirstName: "François", LastName: "Hollande"},
		{FirstName: "Emmanuel", LastName: "Macron"},
	}

	if err = connector.GormDB().Create(&data.politicians).Error; err != nil {
		return
	}

	redactor := &models.User{Email: "redactor@test.com", Tag: "redactor0123"}

	if err = connector.GormDB().Create(redactor).Error; err != nil {
		return
	}

	data.articles = []*models.Article{
		{
			RedactorID:  redactor.ID,
			Title:       "Article about Nicolas Sarkozy",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Nicolas Sarkozy"}},
			Status:      models.ArticleStatusPublished,
		},
		{
			RedactorID:  redactor.ID,
			Title:       "Article about François Hollande",
			Politicians: []*models.Politician{data.politicians[1]},
			Tags:        []*models.ArticleTag{{Tag: "François Hollande"}},
			Status:      models.ArticleStatusPublished,
		}, {
			RedactorID:  redactor.ID,
			Title:       "Article about Emmanuel Macron",
			Politicians: []*models.Politician{data.politicians[2]},
			Tags:        []*models.ArticleTag{{Tag: "Emmanuel Macron"}},
			Status:      models.ArticleStatusPublished,
		},
	}

	err = connector.GormDB().Create(&data.articles).Error

	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := fiberx.NewApp()
	Group(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, Path, nil)

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

	var resDTO []response_dto.Article
	if err = json.Unmarshal(resBody, &resDTO); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(data.articles), len(resDTO))

	for range data.articles {
		assert.True(t, slices.ContainsFunc(resDTO, func(dto response_dto.Article) bool {
			return len(dto.Tags) == 1 && len(dto.Politicians) == 1
		}))
	}
}
