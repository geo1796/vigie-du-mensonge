package redactor_save_article

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
	"vdm/core/dependencies/database"
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
	redactor    *models.User
	politicians []*models.Politician
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

	data.redactor = &models.User{Email: "redactor@test.com", Tag: "redactor0123", Password: "x"}
	if err = connector.GormDB().Create(data.redactor).Error; err != nil {
		return
	}

	data.politicians = []*models.Politician{{FirstName: "Emmanuel", LastName: "Macron"}}
	if err = connector.GormDB().Create(&data.politicians).Error; err != nil {
		return
	}

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

func TestIntegration_NewArticle_PublishTrue_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	longBody := make([]byte, 220)
	for i := range longBody {
		longBody[i] = 'a'
	}
	payload := RequestDTO{
		Title:         "This is a sufficiently long article title",
		EventDate:     time.Now(),
		Category:      models.ArticleCategoryLie,
		Body:          string(longBody),
		Tags:          []string{"Macron"},
		PoliticianIDs: []uuid.UUID{data.politicians[0].ID},
		Sources:       []string{"https://example.com"},
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(Method, Path+"?publish=true", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	// Verify DB state
	var count int64
	if err := connector.GormDB().Model(&models.Article{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), count)

	var article models.Article
	if err := connector.GormDB().Preload("Tags").Preload("Sources").Preload("Politicians").First(&article).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, models.ArticleStatusUnderReview, article.Status)
	assert.Equal(t, int16(1), article.Minor)
	assert.Equal(t, int16(0), article.Major)
	assert.Equal(t, data.redactor.ID, article.RedactorID)
	assert.Equal(t, 1, len(article.Tags))
	assert.Equal(t, 1, len(article.Sources))
	assert.Equal(t, 1, len(article.Politicians))
}

func TestIntegration_NewArticle_PublishFalse_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	payload := RequestDTO{
		Title:         "This is a sufficiently long article title",
		EventDate:     time.Now(),
		Category:      models.ArticleCategoryFalsehood,
		Body:          "", // can be empty for draft
		Tags:          []string{},
		PoliticianIDs: []uuid.UUID{data.politicians[0].ID},
		Sources:       []string{},
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(Method, Path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	var article models.Article
	if err := connector.GormDB().Preload("Tags").First(&article).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, models.ArticleStatusDraft, article.Status)
	assert.Equal(t, int16(0), article.Minor)
	assert.Equal(t, int16(0), article.Major)
	assert.Equal(t, data.redactor.ID, article.RedactorID)
}

func TestIntegration_ExistingArticle_PublishTrue_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	// Seed a draft article
	ref := uuid.New()
	old := &models.Article{
		RedactorID: data.redactor.ID,
		Title:      "Initial draft title that is long enough",
		Body:       "",
		Category:   models.ArticleCategoryLie,
		EventDate:  time.Now(),
		Reference:  ref,
		Status:     models.ArticleStatusDraft,
		Major:      0,
		Minor:      0,
		// no relations to avoid FK on hard delete
	}
	if err := connector.GormDB().Create(old).Error; err != nil {
		t.Fatal(err)
	}

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	longBody := make([]byte, 220)
	for i := range longBody {
		longBody[i] = 'b'
	}
	payload := RequestDTO{
		ID:            old.ID,
		Title:         "Updated title for publish",
		EventDate:     time.Now(),
		Category:      models.ArticleCategoryLie,
		Body:          string(longBody),
		Tags:          []string{"One"},
		PoliticianIDs: []uuid.UUID{data.politicians[0].ID},
		Sources:       []string{"https://source"},
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(Method, Path+"?publish=true", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	// Old draft should be hard-deleted; only one article with this reference should exist now
	var articles []models.Article
	if err := connector.GormDB().Where("reference = ?", ref).Find(&articles).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(articles))
	assert.Equal(t, models.ArticleStatusUnderReview, articles[0].Status)
	assert.Equal(t, int16(1), articles[0].Minor)
	assert.Equal(t, int16(0), articles[0].Major)
}

func TestIntegration_ExistingArticle_PublishFalse_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	// Seed a draft article
	ref := uuid.New()
	old := &models.Article{
		RedactorID: data.redactor.ID,
		Title:      "Draft title",
		Body:       "",
		Category:   models.ArticleCategoryFalsehood,
		EventDate:  time.Now(),
		Reference:  ref,
		Status:     models.ArticleStatusDraft,
		Major:      0,
		Minor:      0,
	}
	if err := connector.GormDB().Create(old).Error; err != nil {
		t.Fatal(err)
	}

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	payload := RequestDTO{
		ID:        old.ID,
		Title:     "Updated draft title which is long enough indeed",
		EventDate: time.Now(),
		Category:  models.ArticleCategoryFalsehood,
		Body:      "updated body but still draft",
		Tags:      []string{"X"},
		Sources:   []string{"https://s1"},
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(Method, Path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	var updated models.Article
	if err := connector.GormDB().First(&updated, "id = ?", old.ID).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, models.ArticleStatusDraft, updated.Status)
	assert.Equal(t, int16(0), updated.Minor)
	assert.Equal(t, int16(0), updated.Major)
}

func TestIntegration_ErrConflict(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	// Seed an article with a status that forbids updates
	ref := uuid.New()
	old := &models.Article{
		RedactorID: data.redactor.ID,
		Title:      "Locked title",
		Body:       "",
		Category:   models.ArticleCategoryLie,
		EventDate:  time.Now(),
		Reference:  ref,
		Status:     models.ArticleStatusUnderReview,
		Major:      0,
		Minor:      1,
	}
	if err := connector.GormDB().Create(old).Error; err != nil {
		t.Fatal(err)
	}

	app := newAppWithAuthedUser(data.redactor.ID)
	Route(connector.GormDB()).Register(app)

	payload := RequestDTO{
		ID:        old.ID,
		Title:     "Attempt update title that is long enough",
		EventDate: time.Now(),
		Category:  models.ArticleCategoryLie,
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(Method, Path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusConflict, res.StatusCode)
}
