package moderator_save_review

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
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
	moderator *models.User
	redactor  *models.User
	article   *models.Article
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
	data.moderator = &models.User{Email: "moderator@test.com", Tag: "moderator0123", Password: "x"}
	if err = connector.GormDB().Create(data.moderator).Error; err != nil {
		return
	}

	ref := uuid.New()
	data.article = &models.Article{
		RedactorID:  data.redactor.ID,
		ModeratorID: &data.moderator.ID,
		Title:       "Pending",
		Status:      models.ArticleStatusUnderReview,
		Reference:   ref,
		Major:       0,
		Minor:       1,
	}
	err = connector.GormDB().Create(data.article).Error
	return
}

func newAppWithAuthedModerator(moderatorID uuid.UUID) *fiber.App {
	app := fiberx.NewApp()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("authedUser", locals.AuthedUser{ID: moderatorID})
		return c.Next()
	})
	return app
}

func TestIntegration_Success_DecisionPublish(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedModerator(data.moderator.ID)
	Route(connector.GormDB()).Register(app)

	payload := map[string]any{"decision": string(models.ArticleStatusPublished)}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(Method, "/"+data.article.ID.String()+"/review", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)

	var updated models.Article
	if err := connector.GormDB().First(&updated, "id = ?", data.article.ID).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, models.ArticleStatusPublished, updated.Status)
	assert.Equal(t, int16(1), updated.Major)
	assert.Equal(t, int16(0), updated.Minor)

	var reviewsCount int64
	if err := connector.GormDB().Model(&models.ArticleReview{}).Where("article_id = ? AND moderator_id = ? AND decision = ?", data.article.ID, data.moderator.ID, models.ArticleStatusPublished).Count(&reviewsCount).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), reviewsCount)
}

func TestIntegration_Success_OtherDecision(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedModerator(data.moderator.ID)
	Route(connector.GormDB()).Register(app)

	notes := strings.Repeat("n", 40)
	payload := map[string]any{"decision": string(models.ArticleStatusChangeRequested), "notes": notes}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(Method, "/"+data.article.ID.String()+"/review", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, res.StatusCode)

	var updated models.Article
	if err := connector.GormDB().First(&updated, "id = ?", data.article.ID).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, models.ArticleStatusChangeRequested, updated.Status)
	assert.Equal(t, int16(0), updated.Major)

	var reviewsCount int64
	if err := connector.GormDB().Model(&models.ArticleReview{}).Where("article_id = ? AND moderator_id = ? AND decision = ? AND notes = ?", data.article.ID, data.moderator.ID, models.ArticleStatusChangeRequested, notes).Count(&reviewsCount).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), reviewsCount)
}
