package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	authpkg "image-to-palette/auth"
	"image-to-palette/db"
	"image-to-palette/model"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"golang.org/x/crypto/bcrypt"
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

func setupPaletteRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/palettes", SavePaletteHandler)
	router.GET("/palettes", GetPalettesHandler)
	return router
}

func setupThemeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/themes", SaveThemeHandler)
	router.PUT("/themes/:id", UpdateThemeHandler)
	router.GET("/themes", GetThemesHandler)
	router.DELETE("/themes/:id", DeleteThemeHandler)
	return router
}

func buildThemePayload(name string, editorType string, signature string) map[string]any {
	return map[string]any{
		"id":                   "local_123",
		"name":                 name,
		"editorType":           editorType,
		"signature":            signature,
		"themeColorsWithUsage": []any{},
		"createdAt":            "2024-01-01T00:00:00Z",
		"themeResult": map[string]any{
			"theme": map[string]any{
				"name": name,
			},
			"themeOverrides": map[string]any{},
			"colors":         []any{},
		},
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	if terminateDBFunc != nil {
		_ = terminateDBFunc(context.Background())
	}
	os.Exit(code)
}

func TestSavePaletteHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupPaletteRouter()

	reqBody, err := json.Marshal(SavePaletteRequest{
		Name:    "My Palette",
		Palette: []model.Color{{Hex: "#FF0000"}},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/palettes", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSavePaletteHandler_Success(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()

	reqBody, err := json.Marshal(SavePaletteRequest{
		Name:    "My Palette",
		Palette: []model.Color{{Hex: "#FF0000"}, {Hex: "#00FF00"}},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/palettes", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var count int64
	if err := db.DB.Model(&model.Palette{}).Count(&count).Error; err != nil {
		t.Fatalf("count palettes: %v", err)
	}
	assert.Equal(t, int64(1), count)
}

func TestGetPalettesHandler_ReturnsSavedPalettes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "Saved", []model.Color{{Hex: "#112233"}}); err != nil {
		t.Fatalf("save palette: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	req := httptest.NewRequest("GET", "/palettes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp GetPalettesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if assert.Len(t, resp.Palettes, 1) {
		assert.Equal(t, "Saved", resp.Palettes[0].Name)
		assert.Equal(t, "#112233", resp.Palettes[0].Palette[0].Hex)
	}
}

func TestSaveThemeHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupThemeRouter()

	body, err := json.Marshal(buildThemePayload("Theme", "vscode", "sig-1"))
	if err != nil {
		t.Fatalf("marshal theme: %v", err)
	}

	req := httptest.NewRequest("POST", "/themes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSaveThemeHandler_CreatesAndDedupes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	payload := buildThemePayload("Theme", "vscode", "sig-1")
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal theme: %v", err)
	}

	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/themes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	first := request()
	assert.Equal(t, http.StatusCreated, first.Code)

	second := request()
	assert.Equal(t, http.StatusOK, second.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(1), count)
}

func TestGetThemesHandler_ReturnsThemes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	_, _, err := saveUserTheme(user.ID, "Theme", "vscode", "sig-1", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	req := httptest.NewRequest("GET", "/themes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp GetThemesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assert.Len(t, resp.Themes, 1)
}

func TestUpdateThemeHandler_UpdatesPayload(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := saveUserTheme(user.ID, "Theme", "vscode", "sig-1", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	payload := buildThemePayload("Theme Updated", "vscode", "sig-2")
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal theme: %v", err)
	}

	req := httptest.NewRequest("PUT", "/themes/"+fmt.Sprintf("%d", saved.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated model.Theme
	if err := db.DB.First(&updated, saved.ID).Error; err != nil {
		t.Fatalf("load updated theme: %v", err)
	}
	assert.Equal(t, "Theme Updated", updated.Name)
	assert.Equal(t, "sig-2", updated.Signature)
}

func TestDeleteThemeHandler_RemovesTheme(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := saveUserTheme(user.ID, "Theme", "vscode", "sig-1", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	req := httptest.NewRequest("DELETE", "/themes/"+fmt.Sprintf("%d", saved.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(0), count)
}
