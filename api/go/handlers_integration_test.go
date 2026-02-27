package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

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

		setupErr = InitDatabase()
	})

	if setupErr != nil {
		t.Fatalf("setup test database: %v", setupErr)
	}
}

func resetTestDB(t *testing.T) {
	t.Helper()
	if DB == nil {
		t.Fatalf("database not initialized")
	}
	if err := DB.Exec("TRUNCATE TABLE palettes, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func createTestUser(t *testing.T) User {
	t.Helper()
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: hash,
	}

	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

func setupPaletteRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/palettes", savePaletteHandler)
	router.GET("/palettes", getPalettesHandler)
	return router
}

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", registerHandler)
	router.POST("/auth/login", loginHandler)
	auth := router.Group("/auth")
	auth.Use(authMiddleware())
	auth.GET("/me", getMeHandler)
	return router
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
		Palette: []Color{{Hex: "#FF0000"}},
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
	token, err := generateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupPaletteRouter()

	reqBody, err := json.Marshal(SavePaletteRequest{
		Name:    "My Palette",
		Palette: []Color{{Hex: "#FF0000"}, {Hex: "#00FF00"}},
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
	if err := DB.Model(&Palette{}).Count(&count).Error; err != nil {
		t.Fatalf("count palettes: %v", err)
	}
	assert.Equal(t, int64(1), count)
}

func TestGetPalettesHandler_ReturnsSavedPalettes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "Saved", []Color{{Hex: "#112233"}}); err != nil {
		t.Fatalf("save palette: %v", err)
	}

	token, err := generateJWTToken(user)
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
	var meBody map[string]User
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
	if err := DB.Model(&user).Update("email", "wrongpass@example.com").Error; err != nil {
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
