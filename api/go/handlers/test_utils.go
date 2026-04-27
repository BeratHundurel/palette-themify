package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	authpkg "themesmith/auth"
	"themesmith/db"
	"themesmith/model"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
)

var (
	setupOnce       sync.Once
	setupErr        error
	terminateDBFunc func(context.Context) error
)

// setupTestDB initializes a test database using testcontainers.
// It runs only once across all tests using sync.Once.
func setupTestDB(t *testing.T) {
	t.Helper()
	setupOnce.Do(func() {
		ctx := context.Background()
		container, err := postgres.Run(
			ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("themesmith_test"),
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
		os.Setenv("DB_NAME", "themesmith_test")
		os.Setenv("DB_SSL_MODE", "disable")
		os.Setenv("JWT_SECRET", "test-secret")

		setupErr = db.InitDatabase()
	})

	if setupErr != nil {
		t.Fatalf("setup test database: %v", setupErr)
	}
}

// resetTestDB truncates all tables to clear state between tests.
func resetTestDB(t *testing.T) {
	t.Helper()
	if db.DB == nil {
		t.Fatalf("database not initialized")
	}
	if err := db.DB.Exec("TRUNCATE TABLE palettes, themes, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

// createTestUser creates a test user with hashed password.
func createTestUser(t *testing.T) model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

// createTestTheme creates a theme for a user.
func createTestTheme(userID uint, name string, editorType string, signature string, jsonData string) (model.Theme, bool, error) {
	theme := model.Theme{
		UserID:     &userID,
		Name:       name,
		EditorType: editorType,
		Signature:  signature,
		JsonData:   datatypes.JSON(jsonData),
	}
	result := db.DB.Create(&theme)
	return theme, result.RowsAffected > 0, result.Error
}

// buildThemePayload builds a theme request payload for testing.
func buildThemePayload(name string, editorType string, signature string) map[string]any {
	return map[string]any{
		"name":       name,
		"editorType": editorType,
		"signature":  signature,
		"themeResult": map[string]any{
			"theme": map[string]any{
				"name": name,
			},
			"themeOverrides": map[string]any{},
			"colors":         []any{},
		},
	}
}

// Router setup functions

// setupPaletteRouter creates a Gin router with palette endpoints.
func setupPaletteRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/palettes", SavePaletteHandler)
	router.GET("/palettes", GetPalettesHandler)
	router.DELETE("/palettes", DeletePalettesBatchHandler)
	return router
}

// setupThemeRouter creates a Gin router with theme endpoints.
func setupThemeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/themes/batch", SaveThemesBatchHandler)
	router.POST("/themes", SaveThemeHandler)
	router.POST("/themes/:id/share", ShareThemeHandler)
	router.DELETE("/themes/:id/share", UnshareThemeHandler)
	router.PUT("/themes/:id", UpdateThemeHandler)
	router.GET("/themes", GetThemesHandler)
	router.DELETE("/themes/:id", DeleteThemeHandler)
	router.DELETE("/themes", DeleteThemesBatchHandler)
	return router
}

// setupSharedRouter creates a Gin router with shared items endpoints.
func setupSharedRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shared-items", GetSharedItemsHandler)
	return router
}

// setupPreferencesRouter creates a Gin router with preferences endpoints.
func setupPreferencesRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authGroup := router.Group("/auth")
	authGroup.Use(authpkg.AuthMiddleware())
	authGroup.GET("/preferences", GetPreferencesHandler)
	authGroup.PUT("/preferences", SavePreferencesHandler)
	return router
}

// generateToken generates a JWT token for a user.
func generateToken(t *testing.T, user model.User) string {
	t.Helper()
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

// makeRequest is a helper to create HTTP requests with auth headers.
func makeRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	return httptest.NewRecorder()
}

// TestMain runs after all tests complete, cleaning up resources.
func TestMain(m *testing.M) {
	code := m.Run()
	if terminateDBFunc != nil {
		_ = terminateDBFunc(context.Background())
	}
	os.Exit(code)
}
