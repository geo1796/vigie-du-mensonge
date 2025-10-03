package moderator_get_claimed_articles

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
	moderator   *models.User
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
	data.moderator = &models.User{Email: "moderator@test.com", Tag: "moderator0123", Password: "x"}
	if err = connector.GormDB().Create(data.moderator).Error; err != nil {
		return
	}

	ref := uuid.New()
	data.articles = []*models.Article{
		{ // should be included
			RedactorID:  data.redactor.ID,
			ModeratorID: &data.moderator.ID,
			Title:       "Claimed pending",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Macron"}},
			Status:      models.ArticleStatusUnderReview,
			Reference:   ref,
		},
		{ // excluded: different moderator
			RedactorID:  data.redactor.ID,
			Title:       "Claimed by someone else",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Macron"}},
			Status:      models.ArticleStatusUnderReview,
			Reference:   ref,
		},
		{ // excluded: wrong status
			RedactorID:  data.redactor.ID,
			ModeratorID: &data.moderator.ID,
			Title:       "Draft",
			Politicians: []*models.Politician{data.politicians[0]},
			Tags:        []*models.ArticleTag{{Tag: "Macron"}},
			Status:      models.ArticleStatusDraft,
			Reference:   ref,
		},
	}
	err = connector.GormDB().Create(&data.articles).Error
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

	req := httptest.NewRequest(Method, Path, nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("Expected status code 200, got %d", res.StatusCode)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var dtos []response_dto.Article
	if err = json.Unmarshal(b, &dtos); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(dtos))
	assert.Equal(t, 1, len(dtos[0].Politicians))
	assert.Equal(t, 1, len(dtos[0].Tags))
}
