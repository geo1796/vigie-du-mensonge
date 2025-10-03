package redactor_get_articles

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/dto/response_dto"
	"vdm/core/fiberx"
	"vdm/core/locals"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	politicians []*models.Politician
	redactor    *models.User
	other       *models.User
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

	data.redactor = &models.User{Email: "redactor@test.com", Tag: "redactor0123", Password: "x"}
	if err = connector.GormDB().Create(data.redactor).Error; err != nil {
		return
	}
	data.other = &models.User{Email: "other@test.com", Tag: "other0123", Password: "x"}
	if err = connector.GormDB().Create(data.other).Error; err != nil {
		return
	}

	ref := uuid.New()

	data.articles = []*models.Article{
		{
			RedactorID:  data.redactor.ID,
			Title:       "Article about Nicolas Sarkozy",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Nicolas Sarkozy"}},
			Status:      models.ArticleStatusPublished,
			Reference:   ref,
		},
		{
			RedactorID:  data.redactor.ID,
			Title:       "Article about François Hollande",
			Politicians: []*models.Politician{data.politicians[1]},
			Tags:        []*models.ArticleTag{{Tag: "François Hollande"}},
			Status:      models.ArticleStatusDraft,
			Reference:   ref,
		},
		{
			RedactorID:  data.redactor.ID,
			Title:       "Archived article to be excluded",
			Politicians: []*models.Politician{data.politicians[2]},
			Tags:        []*models.ArticleTag{{Tag: "Excluded"}},
			Status:      models.ArticleStatusArchived,
			Reference:   ref,
		},
		{
			RedactorID:  data.other.ID,
			Title:       "Other redactor article",
			Politicians: []*models.Politician{data.politicians[2]},
			Tags:        []*models.ArticleTag{{Tag: "Other"}},
			Status:      models.ArticleStatusPublished,
			Reference:   ref,
		},
	}

	err = connector.GormDB().Create(&data.articles).Error
	return
}

func newAppWithAuthedUser(redactorID uuid.UUID) *fiber.App {
	app := fiberx.NewApp()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("authedUser", locals.AuthedUser{ID: redactorID})
		return c.Next()
	})
	return app
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

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

	// Expect only 2 articles for that redactor that are not archived
	assert.Equal(t, 2, len(resDTO))
	for i := range resDTO {
		assert.Equal(t, 1, len(resDTO[i].Politicians))
		assert.Equal(t, 1, len(resDTO[i].Tags))
	}
}
