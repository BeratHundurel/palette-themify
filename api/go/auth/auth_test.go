package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"image-to-palette/db"
	"image-to-palette/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	setupOnce       sync.Once
	setupErr        error
	terminateDBFunc func(context.Context) error
)

func setupTestDB(t *testing.T) {
	t.Helper()
	setupOnce.Do(func() {
		ctx := context.Background()
		container, err := postgres.Run(
			ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("image_to_palette_test"),
			postgres.WithUsername("postgres"),
			postgres.WithPassword("password"),
			postgres.BasicWaitStrategies(),
		)
		if err != nil {
			setupErr = err
			return
		}

		terminateDBFunc = func(ctx context.Context) error {
			return container.Terminate(ctx)
		}

		host, err := container.Host(ctx)
		if err != nil {
			setupErr = err
			return
		}
		port, err := container.MappedPort(ctx, "5432/tcp")
		if err != nil {
			setupErr = err
			return
		}

		os.Setenv("DB_HOST", host)
		os.Setenv("DB_PORT", port.Port())
		os.Setenv("DB_USER", "postgres")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_NAME", "image_to_palette_test")
		os.Setenv("DB_SSL_MODE", "disable")
		os.Setenv("JWT_SECRET", "test-secret")

		setupErr = db.InitDatabase()
	})

	if setupErr != nil {
		t.Fatalf("setup test database: %v", setupErr)
	}
}

func resetTestDB(t *testing.T) {
	t.Helper()
	if db.DB == nil {
		t.Fatalf("database not initialized")
	}
	if err := db.DB.Exec("TRUNCATE TABLE palettes, themes, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func createTestUser(t *testing.T) model.User {
	t.Helper()
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: hash,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", RegisterHandler)
	router.POST("/auth/login", LoginHandler)
	authGroup := router.Group("/auth")
	authGroup.Use(AuthMiddleware())
	authGroup.GET("/me", GetMeHandler)
	return router
}

func TestMain(m *testing.M) {
	code := m.Run()
	if terminateDBFunc != nil {
		_ = terminateDBFunc(context.Background())
	}
	os.Exit(code)
}

func TestAuthRegisterLoginMeFlow(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupAuthRouter()

	registerBody, err := json.Marshal(RegisterRequest{
		Name:     "Test User",
		Email:    "auth@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("marshal register: %v", err)
	}

	registerReq := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	assert.Equal(t, http.StatusCreated, registerResp.Code)

	var registerAuth AuthResponse
	if err := json.Unmarshal(registerResp.Body.Bytes(), &registerAuth); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registerAuth.Token == "" {
		t.Fatalf("missing register token")
	}

	meReq := httptest.NewRequest("GET", "/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+registerAuth.Token)
	meResp := httptest.NewRecorder()
	router.ServeHTTP(meResp, meReq)

	assert.Equal(t, http.StatusOK, meResp.Code)
	var meBody map[string]model.User
	if err := json.Unmarshal(meResp.Body.Bytes(), &meBody); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if user, ok := meBody["user"]; ok {
		assert.Equal(t, "auth@example.com", user.Email)
	}

	loginBody, err := json.Marshal(LoginRequest{
		Email:    "auth@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("marshal login: %v", err)
	}

	loginReq := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	assert.Equal(t, http.StatusOK, loginResp.Code)
	var loginAuth AuthResponse
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginAuth); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginAuth.Token == "" {
		t.Fatalf("missing login token")
	}
}

func TestAuthLoginInvalidPassword(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := db.DB.Model(&user).Update("email", "wrongpass@example.com").Error; err != nil {
		t.Fatalf("update user email: %v", err)
	}

	router := setupAuthRouter()
	loginBody, err := json.Marshal(LoginRequest{
		Email:    "wrongpass@example.com",
		Password: "not-the-right-password",
	})
	if err != nil {
		t.Fatalf("marshal login: %v", err)
	}

	loginReq := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	assert.Equal(t, http.StatusUnauthorized, loginResp.Code)
}
