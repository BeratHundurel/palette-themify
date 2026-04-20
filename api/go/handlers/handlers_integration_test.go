package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	authpkg "themesmith/auth"
	"themesmith/db"
	"themesmith/model"

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
	router.DELETE("/palettes", DeletePalettesBatchHandler)
	return router
}

func setupThemeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/themes", SaveThemeHandler)
	router.POST("/themes/:id/share", ShareThemeHandler)
	router.DELETE("/themes/:id/share", UnshareThemeHandler)
	router.PUT("/themes/:id", UpdateThemeHandler)
	router.GET("/themes", GetThemesHandler)
	router.DELETE("/themes/:id", DeleteThemeHandler)
	router.DELETE("/themes", DeleteThemesBatchHandler)
	return router
}

func setupSharedRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shared-items", GetSharedItemsHandler)
	return router
}

func setupPreferencesRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authGroup := router.Group("/auth")
	authGroup.Use(authpkg.AuthMiddleware())
	authGroup.GET("/preferences", GetPreferencesHandler)
	authGroup.PUT("/preferences", SavePreferencesHandler)
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

func TestSaveThemeHandler_DedupesLongSignature(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	longSignature := strings.Repeat("sig-very-long-", 20)
	payload := buildThemePayload("Theme", "vscode", longSignature)
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

func TestDeleteThemesBatchHandler_RemovesOnlyRequestedThemes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	theme1, _, err := saveUserTheme(user.ID, "Theme 1", "vscode", "sig-1", `{"name":"Theme 1"}`)
	if err != nil {
		t.Fatalf("save theme1: %v", err)
	}
	theme2, _, err := saveUserTheme(user.ID, "Theme 2", "vscode", "sig-2", `{"name":"Theme 2"}`)
	if err != nil {
		t.Fatalf("save theme2: %v", err)
	}
	_, _, err = saveUserTheme(user.ID, "Theme 3", "vscode", "sig-3", `{"name":"Theme 3"}`)
	if err != nil {
		t.Fatalf("save theme3: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	body, err := json.Marshal(map[string]any{
		"ids": []string{fmt.Sprintf("%d", theme1.ID), fmt.Sprintf("%d", theme2.ID)},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/themes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(1), count)
}

func TestDeleteThemesBatchHandler_InvalidPayload(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	req := httptest.NewRequest("DELETE", "/themes", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteThemesBatchHandler_EmptyIDs(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	body, err := json.Marshal(map[string]any{"ids": []string{}})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/themes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteThemesBatchHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupThemeRouter()
	body, err := json.Marshal(map[string]any{"ids": []string{"1"}})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/themes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteThemesBatchHandler_IgnoresOtherUsersIDs(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	owner := createTestUser(t)
	ownerTheme, _, err := saveUserTheme(owner.ID, "Owner Theme", "vscode", "owner-sig", `{"name":"Owner Theme"}`)
	if err != nil {
		t.Fatalf("save owner theme: %v", err)
	}

	otherUser := model.User{Name: "Other User", Email: "other@example.com", PasswordHash: "hash"}
	if err := db.DB.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}
	otherTheme, _, err := saveUserTheme(otherUser.ID, "Other Theme", "vscode", "other-sig", `{"name":"Other Theme"}`)
	if err != nil {
		t.Fatalf("save other theme: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(owner)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	body, err := json.Marshal(map[string]any{
		"ids": []string{fmt.Sprintf("%d", ownerTheme.ID), fmt.Sprintf("%d", otherTheme.ID)},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/themes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ownerCount int64
	if err := db.DB.Model(&model.Theme{}).Where("id = ?", ownerTheme.ID).Count(&ownerCount).Error; err != nil {
		t.Fatalf("count owner theme: %v", err)
	}
	assert.Equal(t, int64(0), ownerCount)

	var otherCount int64
	if err := db.DB.Model(&model.Theme{}).Where("id = ?", otherTheme.ID).Count(&otherCount).Error; err != nil {
		t.Fatalf("count other theme: %v", err)
	}
	assert.Equal(t, int64(1), otherCount)
}

func TestDeletePaletteHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupPaletteRouter()
	router.DELETE("/palettes/:id", DeletePaletteHandler)

	req := httptest.NewRequest("DELETE", "/palettes/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeletePaletteHandler_Success(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "To Delete", []model.Color{{Hex: "#FF0000"}}); err != nil {
		t.Fatalf("save palette: %v", err)
	}

	var saved model.Palette
	if err := db.DB.Where("user_id = ?", user.ID).First(&saved).Error; err != nil {
		t.Fatalf("load saved palette: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	router.DELETE("/palettes/:id", DeletePaletteHandler)

	req := httptest.NewRequest("DELETE", "/palettes/"+fmt.Sprintf("%d", saved.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	if err := db.DB.Model(&model.Palette{}).Count(&count).Error; err != nil {
		t.Fatalf("count palettes: %v", err)
	}
	assert.Equal(t, int64(0), count)
}

func TestDeletePaletteHandler_InvalidID(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	router.DELETE("/palettes/:id", DeletePaletteHandler)

	req := httptest.NewRequest("DELETE", "/palettes/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeletePaletteHandler_CannotDeleteOtherUsersPalette(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user1 := createTestUser(t)
	if err := saveUserPalette(user1.ID, "User1 Palette", []model.Color{{Hex: "#FF0000"}}); err != nil {
		t.Fatalf("save palette: %v", err)
	}

	var palette model.Palette
	if err := db.DB.Where("user_id = ?", user1.ID).First(&palette).Error; err != nil {
		t.Fatalf("load palette: %v", err)
	}

	user2 := model.User{
		Name:         "User 2",
		Email:        "user2@example.com",
		PasswordHash: "hash",
	}
	if err := db.DB.Create(&user2).Error; err != nil {
		t.Fatalf("create user2: %v", err)
	}

	token2, err := authpkg.GenerateJWTToken(user2)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	router.DELETE("/palettes/:id", DeletePaletteHandler)

	req := httptest.NewRequest("DELETE", "/palettes/"+fmt.Sprintf("%d", palette.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var count int64
	if err := db.DB.Model(&model.Palette{}).Count(&count).Error; err != nil {
		t.Fatalf("count palettes: %v", err)
	}
	assert.Equal(t, int64(1), count, "Palette should not be deleted")
}

func TestDeletePalettesBatchHandler_RemovesOnlyNonSystemRequested(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "Palette 1", []model.Color{{Hex: "#FF0000"}}); err != nil {
		t.Fatalf("save palette1: %v", err)
	}
	if err := saveUserPalette(user.ID, "Palette 2", []model.Color{{Hex: "#00FF00"}}); err != nil {
		t.Fatalf("save palette2: %v", err)
	}

	var palettes []model.Palette
	if err := db.DB.Where("user_id = ?", user.ID).Order("id ASC").Find(&palettes).Error; err != nil {
		t.Fatalf("load palettes: %v", err)
	}
	if len(palettes) < 2 {
		t.Fatalf("expected at least 2 palettes, got %d", len(palettes))
	}

	if err := db.DB.Model(&model.Palette{}).Where("id = ?", palettes[1].ID).Update("is_system", true).Error; err != nil {
		t.Fatalf("mark system palette: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	body, err := json.Marshal(map[string]any{
		"ids": []string{fmt.Sprintf("%d", palettes[0].ID), fmt.Sprintf("%d", palettes[1].ID)},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/palettes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var remainingCount int64
	if err := db.DB.Model(&model.Palette{}).Where("user_id = ?", user.ID).Count(&remainingCount).Error; err != nil {
		t.Fatalf("count remaining palettes: %v", err)
	}
	assert.Equal(t, int64(1), remainingCount)

	var systemStillExists int64
	if err := db.DB.Model(&model.Palette{}).Where("id = ? AND is_system = ?", palettes[1].ID, true).Count(&systemStillExists).Error; err != nil {
		t.Fatalf("count system palette: %v", err)
	}
	assert.Equal(t, int64(1), systemStillExists)
}

func TestDeletePalettesBatchHandler_InvalidPayload(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	req := httptest.NewRequest("DELETE", "/palettes", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePalettesBatchHandler_EmptyIDs(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	body, err := json.Marshal(map[string]any{"ids": []string{}})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/palettes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePalettesBatchHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupPaletteRouter()
	body, err := json.Marshal(map[string]any{"ids": []string{"1"}})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/palettes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeletePalettesBatchHandler_IgnoresOtherUsersIDs(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	owner := createTestUser(t)
	if err := saveUserPalette(owner.ID, "Owner Palette", []model.Color{{Hex: "#AA0000"}}); err != nil {
		t.Fatalf("save owner palette: %v", err)
	}

	var ownerPalette model.Palette
	if err := db.DB.Where("user_id = ?", owner.ID).First(&ownerPalette).Error; err != nil {
		t.Fatalf("load owner palette: %v", err)
	}

	otherUser := model.User{Name: "Palette Other", Email: "palette-other@example.com", PasswordHash: "hash"}
	if err := db.DB.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}
	if err := saveUserPalette(otherUser.ID, "Other Palette", []model.Color{{Hex: "#00AA00"}}); err != nil {
		t.Fatalf("save other palette: %v", err)
	}

	var otherPalette model.Palette
	if err := db.DB.Where("user_id = ?", otherUser.ID).First(&otherPalette).Error; err != nil {
		t.Fatalf("load other palette: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(owner)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()
	body, err := json.Marshal(map[string]any{
		"ids": []string{fmt.Sprintf("%d", ownerPalette.ID), fmt.Sprintf("%d", otherPalette.ID)},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/palettes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ownerCount int64
	if err := db.DB.Model(&model.Palette{}).Where("id = ?", ownerPalette.ID).Count(&ownerCount).Error; err != nil {
		t.Fatalf("count owner palette: %v", err)
	}
	assert.Equal(t, int64(0), ownerCount)

	var otherCount int64
	if err := db.DB.Model(&model.Palette{}).Where("id = ?", otherPalette.ID).Count(&otherCount).Error; err != nil {
		t.Fatalf("count other palette: %v", err)
	}
	assert.Equal(t, int64(1), otherCount)
}

func TestDeleteThemeHandler_CannotDeleteOtherUsersTheme(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user1 := createTestUser(t)
	theme1, _, err := saveUserTheme(user1.ID, "User1 Theme", "vscode", "sig-1", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	user2 := model.User{
		Name:         "User 2",
		Email:        "user2@example.com",
		PasswordHash: "hash",
	}
	if err := db.DB.Create(&user2).Error; err != nil {
		t.Fatalf("create user2: %v", err)
	}

	token2, err := authpkg.GenerateJWTToken(user2)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	req := httptest.NewRequest("DELETE", "/themes/"+fmt.Sprintf("%d", theme1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(1), count, "Theme should not be deleted")
}

func TestGetPreferencesHandler_NoPreferences(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPreferencesRouter()

	req := httptest.NewRequest("GET", "/auth/preferences", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assert.Nil(t, resp["preferences"])
}

func TestSavePreferencesHandler_CreateNew(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPreferencesRouter()

	prefsPayload := map[string]any{
		"theme":    "dark",
		"language": "en",
	}
	body, err := json.Marshal(prefsPayload)
	if err != nil {
		t.Fatalf("marshal preferences: %v", err)
	}

	req := httptest.NewRequest("PUT", "/auth/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PreferencesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	var savedPrefs map[string]any
	if err := json.Unmarshal(resp.Preferences, &savedPrefs); err != nil {
		t.Fatalf("decode preferences: %v", err)
	}
	assert.Equal(t, "dark", savedPrefs["theme"])
	assert.Equal(t, "en", savedPrefs["language"])
}

func TestSavePreferencesHandler_UpdateExisting(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	initialPrefs := model.UserPreferences{
		UserID:   user.ID,
		JsonData: `{"theme":"light"}`,
	}
	if err := db.DB.Create(&initialPrefs).Error; err != nil {
		t.Fatalf("create initial preferences: %v", err)
	}

	router := setupPreferencesRouter()

	updatedPayload := map[string]any{
		"theme":    "dark",
		"language": "es",
	}
	body, err := json.Marshal(updatedPayload)
	if err != nil {
		t.Fatalf("marshal preferences: %v", err)
	}

	req := httptest.NewRequest("PUT", "/auth/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PreferencesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	var savedPrefs map[string]any
	if err := json.Unmarshal(resp.Preferences, &savedPrefs); err != nil {
		t.Fatalf("decode preferences: %v", err)
	}
	assert.Equal(t, "dark", savedPrefs["theme"])
	assert.Equal(t, "es", savedPrefs["language"])
}

func TestGetPreferencesHandler_ReturnsExisting(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	prefs := model.UserPreferences{
		UserID:   user.ID,
		JsonData: `{"theme":"dark","language":"fr"}`,
	}
	if err := db.DB.Create(&prefs).Error; err != nil {
		t.Fatalf("create preferences: %v", err)
	}

	router := setupPreferencesRouter()

	req := httptest.NewRequest("GET", "/auth/preferences", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PreferencesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	var savedPrefs map[string]any
	if err := json.Unmarshal(resp.Preferences, &savedPrefs); err != nil {
		t.Fatalf("decode preferences: %v", err)
	}
	assert.Equal(t, "dark", savedPrefs["theme"])
	assert.Equal(t, "fr", savedPrefs["language"])
}

func TestPreferencesHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupPreferencesRouter()

	t.Run("GetRequiresAuth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/preferences", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("SaveRequiresAuth", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"theme": "dark"})
		req := httptest.NewRequest("PUT", "/auth/preferences", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestSavePaletteHandler_InvalidJSON(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()

	req := httptest.NewRequest("POST", "/palettes", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSaveThemeHandler_InvalidJSON(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	req := httptest.NewRequest("POST", "/themes", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateThemeHandler_NotFound(t *testing.T) {
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

	req := httptest.NewRequest("PUT", "/themes/99999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestShareThemeHandler_SetsSharedFlags(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := saveUserTheme(user.ID, "Theme", "vscode", "sig-share", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	req := httptest.NewRequest("POST", "/themes/"+fmt.Sprintf("%d", saved.ID)+"/share", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated model.Theme
	if err := db.DB.First(&updated, saved.ID).Error; err != nil {
		t.Fatalf("load theme: %v", err)
	}
	assert.True(t, updated.IsShared)
	assert.NotNil(t, updated.SharedAt)
}

func TestUnshareThemeHandler_ClearsSharedFlags(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := saveUserTheme(user.ID, "Theme", "vscode", "sig-unshare", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	if _, err := setThemeShared(user.ID, fmt.Sprintf("%d", saved.ID), true); err != nil {
		t.Fatalf("share theme: %v", err)
	}

	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()
	req := httptest.NewRequest("DELETE", "/themes/"+fmt.Sprintf("%d", saved.ID)+"/share", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated model.Theme
	if err := db.DB.First(&updated, saved.ID).Error; err != nil {
		t.Fatalf("load theme: %v", err)
	}
	assert.False(t, updated.IsShared)
	assert.Nil(t, updated.SharedAt)
}

func TestGetSharedItemsHandler_ReturnsSharedPalettesAndThemes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "Shared Palette", []model.Color{{Hex: "#112233"}, {Hex: "#445566"}}); err != nil {
		t.Fatalf("save palette: %v", err)
	}

	var savedPalette model.Palette
	if err := db.DB.Where("user_id = ?", user.ID).First(&savedPalette).Error; err != nil {
		t.Fatalf("load palette: %v", err)
	}
	if _, err := setPaletteShared(user.ID, fmt.Sprintf("%d", savedPalette.ID), true); err != nil {
		t.Fatalf("share palette: %v", err)
	}

	_, _, err := saveUserTheme(
		user.ID,
		"Shared Theme",
		"vscode",
		"sig-shared-items",
		`{"name":"Shared Theme","themeResult":{"theme":{"name":"Shared Theme"},"colors":[{"hex":"#AABBCC"}]}}`,
	)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	var savedTheme model.Theme
	if err := db.DB.Where("user_id = ?", user.ID).Order("id DESC").First(&savedTheme).Error; err != nil {
		t.Fatalf("load theme: %v", err)
	}
	if _, err := setThemeShared(user.ID, fmt.Sprintf("%d", savedTheme.ID), true); err != nil {
		t.Fatalf("share theme: %v", err)
	}

	router := setupSharedRouter()
	req := httptest.NewRequest("GET", "/shared-items?q=Shared&sort=name", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SharedItemsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "Shared Palette", resp.Items[0].Name)
	assert.Equal(t, SharedItemKindPalette, resp.Items[0].Kind)
	assert.Equal(t, "Shared Theme", resp.Items[1].Name)
	assert.Equal(t, SharedItemKindTheme, resp.Items[1].Kind)
}
