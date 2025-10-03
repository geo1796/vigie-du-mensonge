package moderator_claim_article

import (
	"context"
	"net/http/httptest"
	"testing"
	"vdm/core/dependencies/database"
	"vdm/core/fiberx"
	"vdm/core/locals"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

type testData struct {
	politicians []*models.Politician
	redactor    *models.User
	moderator   *models.User
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
		Title:       "Pending",
		Politicians: []*models.Politician{data.politicians[0]},
		Tags:        []*models.ArticleTag{{Tag: "Macron"}},
		Status:      models.ArticleStatusUnderReview,
		Reference:   ref,
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

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, connector, data := loadTestData(c, t)
	t.Cleanup(func() { test_utils.CleanUpTestData(c, t, container, connector) })

	app := newAppWithAuthedModerator(data.moderator.ID)
	Route(connector.GormDB()).Register(app)

	req := httptest.NewRequest(Method, "/"+data.article.ID.String()+"/claim", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}

	var cnt int64
	if err := connector.GormDB().Model(&models.Article{}).
		Where("id = ? AND moderator_id = ?", data.article.ID, data.moderator.ID).
		Count(&cnt).Error; err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Fatalf("expected article to be claimed by moderator")
	}
}
