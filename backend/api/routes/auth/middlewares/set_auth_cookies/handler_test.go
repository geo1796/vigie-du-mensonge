package set_auth_cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
	"vdm/core/locals"
	"vdm/core/locals/local_keys"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Success(t *testing.T) {
	app := fiberx.NewApp()

	dummyCfg := env.SecurityConfig{
		AccessTokenSecret: []byte("dummySecret"),
		AccessTokenTTL:    1 * time.Minute,
		RefreshTokenTTL:   1 * time.Minute,
		AccessCookieName:  "jwt",
		RefreshCookieName: "rft",
		CookieSecure:      true,
		CookieSameSite:    "strict",
	}

	Middleware(dummyCfg).Register(app)

	accessToken := locals.AccessToken{Token: "jwt", Expiry: time.Now()}
	refreshToken := locals.RefreshToken{Token: uuid.New(), Expiry: time.Now()}

	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals(local_keys.AccessToken, accessToken)
		c.Locals(local_keys.RefreshToken, refreshToken)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)
	assert.Equal(t, accessToken.Token, res.Cookies()[0].Value)
	assert.Equal(t, http.SameSiteStrictMode, res.Cookies()[0].SameSite)
	assert.Equal(t, refreshToken.Token.String(), res.Cookies()[1].Value)
	assert.Equal(t, http.SameSiteStrictMode, res.Cookies()[1].SameSite)
}
