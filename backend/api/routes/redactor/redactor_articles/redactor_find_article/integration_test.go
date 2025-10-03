package redactor_find_article

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
	ref         uuid.UUID
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

	data.politicians = []*models.Politician{{FirstName: "Emmanuel", LastName: "Macron"}}
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

	data.ref = uuid.New()

	data.articles = []*models.Article{
		{
			RedactorID:  data.redactor.ID,
			Title:       "Article v1",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Macron"}},
			Status:      models.ArticleStatusDraft,
			Reference:   data.ref,
		},
		{
			RedactorID:  data.redactor.ID,
			Title:       "Article v2",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Macron"}},
			Status:      models.ArticleStatusUnderReview,
			Reference:   data.ref,
		},
		{
			RedactorID:  data.other.ID,
			Title:       "Other's article same ref",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Macron"}},
			Status:      models.ArticleStatusDraft,
			Reference:   data.ref,
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

	req := httptest.NewRequest(Method, "/"+data.ref.String(), nil)

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

	// Only the two articles belonging to the redactor with the same reference should be returned
	assert.Equal(t, 2, len(resDTO))
	for i := range resDTO {
		assert.Equal(t, 1, len(resDTO[i].Politicians))
		assert.Equal(t, 1, len(resDTO[i].Tags))
	}
}

func TestIntegration_ErrNotFound(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedUser(data.redactor.ID)
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
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/not-a-uuid", nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)
}

func TestIntegration_WrongRedactor(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	// Create an article for the other redactor with a unique reference
	otherRef := uuid.New()
	extra := &models.Article{
		RedactorID:  data.other.ID,
		Title:       "Other redactor unique ref",
		Politicians: []*models.Politician{data.politicians[0]},
		Tags:        []*models.ArticleTag{{Tag: "Other"}},
		Status:      models.ArticleStatusDraft,
		Reference:   otherRef,
	}
	if err := connector.GormDB().Create(extra).Error; err != nil {
		t.Fatal(err)
	}

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/"+otherRef.String(), nil)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	// The authed redactor should not see other redactor's articles, expect 404
	assert.Equal(t, fiber.StatusNotFound, res.StatusCode)
}
