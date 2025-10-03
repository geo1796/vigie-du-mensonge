package sign_in

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/fiberx"
	"vdm/core/models"
	"vdm/test_utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/bcrypt"
)

var testRoles = []*models.Role{
	{Name: "ADMIN"},
	{Name: "MODERATOR"},
}
var testUser = &models.User{Email: "signin_user0@email.com", Tag: "user0123", Roles: testRoles}

func loadPgxData(c context.Context, t *testing.T) (container testcontainers.Container, pgxProvider database.PgxProvider) {
	container, pgxProvider = test_utils.NewTestContainerPgxProvider(c, t)

	pool := pgxProvider.Pool()
	var err error

	defer func() {
		if err != nil {
			test_utils.CleanUpPgxProvider(c, t, container, pgxProvider)
			t.Fatal(err)
		}
	}()

	// create a user with known password
	pwd, _ := bcrypt.GenerateFromPassword([]byte("Test123!"), bcrypt.DefaultCost)
	testUser.Password = string(pwd)

	// Insert roles and fetch their IDs (upsert-safe)
	roleIDs := make([]string, len(testRoles))
	for i, r := range testRoles {
		row := pool.QueryRow(c, `INSERT INTO roles (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id`, r.Name)
		if scanErr := row.Scan(&roleIDs[i]); scanErr != nil {
			err = scanErr
			return
		}
	}

	// Insert user and get ID (upsert by email/tag for safety)
	var userID string
	row := pool.QueryRow(c, `INSERT INTO users (tag, email, password)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET tag = EXCLUDED.tag, password = EXCLUDED.password
		RETURNING id`, testUser.Tag, testUser.Email, testUser.Password)
	if scanErr := row.Scan(&userID); scanErr != nil {
		err = scanErr
		return
	}

	// Map user to roles
	for _, roleID := range roleIDs {
		_, execErr := pool.Exec(c, `INSERT INTO user_roles (user_id, role_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, role_id) DO NOTHING`, userID, roleID)
		if execErr != nil {
			err = execErr
			return
		}
	}

	return
}

func TestIntegration_Success(t *testing.T) {
	c := context.Background()
	container, pgxProvider := loadPgxData(c, t)
	t.Cleanup(func() { test_utils.CleanUpPgxProvider(c, t, container, pgxProvider) })

	app := fiberx.NewApp()

	dummyCfg := env.SecurityConfig{AccessTokenSecret: []byte("dummySecret"), AccessTokenTTL: 1 * time.Minute, RefreshTokenTTL: 1 * time.Minute}

	Route(pgxProvider, dummyCfg).Register(app)

	reqDTO := RequestDTO{
		Email:    testUser.Email,
		Password: "Test123!",
	}
	b, _ := json.Marshal(reqDTO)

	req := httptest.NewRequest(Method, Path, bytes.NewReader(b))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	// parse body and assert roles length
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var dto ResponseDTO
	if err := json.Unmarshal(body, &dto); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testUser.RoleNames(), dto.Roles)

	var count int64
	if err = pgxProvider.Pool().
		QueryRow(c, "SELECT COUNT(*) FROM user_tokens").
		Scan(&count); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(1), count)
}

func TestIntegration_WrongPassword(t *testing.T) {
	c := context.Background()
	container, pgxProvider := loadPgxData(c, t)
	t.Cleanup(func() { test_utils.CleanUpPgxProvider(c, t, container, pgxProvider) })

	app := fiberx.NewApp()

	dummyCfg := env.SecurityConfig{AccessTokenSecret: []byte("dummySecret"), AccessTokenTTL: 1 * time.Minute, RefreshTokenTTL: 1 * time.Minute}

	Route(pgxProvider, dummyCfg).Register(app)

	reqDTO := RequestDTO{
		Email:    testUser.Email,
		Password: "WrongPassword" + time.Now().String(),
	}
	b, _ := json.Marshal(reqDTO)

	req := httptest.NewRequest(Method, Path, bytes.NewReader(b))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, res.StatusCode)
}

func TestIntegration_EmailNotFound(t *testing.T) {
	c := context.Background()
	container, pgxProvider := loadPgxData(c, t)
	t.Cleanup(func() { test_utils.CleanUpPgxProvider(c, t, container, pgxProvider) })

	app := fiberx.NewApp()

	dummyCfg := env.SecurityConfig{AccessTokenSecret: []byte("dummySecret"), AccessTokenTTL: 1 * time.Minute, RefreshTokenTTL: 1 * time.Minute}

	Route(pgxProvider, dummyCfg).Register(app)

	reqDTO := RequestDTO{
		Email:    "unknown_" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@email.com",
		Password: "SomePassword1!",
	}
	b, _ := json.Marshal(reqDTO)

	req := httptest.NewRequest(Method, Path, bytes.NewReader(b))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, res.StatusCode)
}
