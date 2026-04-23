package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"themesmith/db"
	"themesmith/model"

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
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := model.User{
		Name:         "Test User",
		Email:        "test-user@local.themesmith",
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
	router.GET("/auth/google", GoogleLoginHandler)
	router.GET("/auth/google/callback", GoogleCallbackHandler)
	router.GET("/auth/google/exchange", GoogleExchangeCodeHandler)
	router.GET("/auth/google/desktop/status", GoogleDesktopStatusHandler)
	authGroup := router.Group("/auth")
	authGroup.Use(AuthMiddleware())
	authGroup.GET("/me", GetMeHandler)
	authGroup.POST("/change-password", ChangePasswordHandler)
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
		Username: "auth-user",
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
		assert.Equal(t, "auth-user", user.Name)
	}

	loginBody, err := json.Marshal(LoginRequest{
		Username: "auth-user",
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
	if err := db.DB.Model(&user).Update("name", "wrongpassuser").Error; err != nil {
		t.Fatalf("update user username: %v", err)
	}

	router := setupAuthRouter()
	loginBody, err := json.Marshal(LoginRequest{
		Username: "wrongpassuser",
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

func TestChangePasswordHandler_Success(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupAuthRouter()

	changePassBody, err := json.Marshal(map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword456",
	})
	if err != nil {
		t.Fatalf("marshal change password request: %v", err)
	}

	req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(changePassBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedUser model.User
	if err := db.DB.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}

	assert.True(t, CheckPasswordHash("newpassword456", updatedUser.PasswordHash))
}

func TestChangePasswordHandler_WrongCurrentPassword(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupAuthRouter()

	changePassBody, err := json.Marshal(map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword456",
	})
	if err != nil {
		t.Fatalf("marshal change password request: %v", err)
	}

	req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(changePassBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestChangePasswordHandler_InvalidPayload(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupAuthRouter()

	t.Run("MissingNewPassword", func(t *testing.T) {
		changePassBody, _ := json.Marshal(map[string]string{
			"current_password": "password123",
		})

		req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(changePassBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("TooShortNewPassword", func(t *testing.T) {
		changePassBody, _ := json.Marshal(map[string]string{
			"current_password": "password123",
			"new_password":     "short",
		})

		req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(changePassBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestChangePasswordHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupAuthRouter()

	changePassBody, _ := json.Marshal(map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword456",
	})

	req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(changePassBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupAuthRouter()

	t.Run("MalformedToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-format")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("MissingBearerPrefix", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "some-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("NoAuthorizationHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("TokenForNonexistentUser", func(t *testing.T) {
		nonExistentUser := model.User{
			ID:    99999,
			Name:  "Nonexistent",
			Email: "nonexistent@example.com",
		}
		token, err := GenerateJWTToken(nonExistentUser)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}

		req := httptest.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGoogleExchangeCodeHandler_Success(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupAuthRouter()
	user := createTestUser(t)

	googleAuthSessionsMu.Lock()
	googleAuthSessions = map[string]*googleAuthSession{}
	googleAuthSessions["web-session"] = &googleAuthSession{
		Mode:         googleAuthModeWeb,
		RedirectURL:  "http://localhost:5173/auth/google/callback",
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		Token:        "jwt-token",
		User:         &user,
		ExchangeCode: "exchange-code-123",
	}
	googleAuthSessionsMu.Unlock()

	req := httptest.NewRequest("GET", "/auth/google/exchange?exchange_code=exchange-code-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var authResponse AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &authResponse); err != nil {
		t.Fatalf("decode exchange response: %v", err)
	}

	assert.Equal(t, "jwt-token", authResponse.Token)
	assert.Equal(t, user.ID, authResponse.User.ID)
	assert.Equal(t, "Login successful", authResponse.Message)

	googleAuthSessionsMu.Lock()
	_, exists := googleAuthSessions["web-session"]
	googleAuthSessionsMu.Unlock()
	assert.False(t, exists)
}

func TestGoogleExchangeCodeHandler_InvalidOrExpiredCode(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupAuthRouter()

	t.Run("MissingExchangeCode", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/google/exchange", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Missing exchange_code parameter"}`, w.Body.String())
	})

	t.Run("UnknownExchangeCode", func(t *testing.T) {
		googleAuthSessionsMu.Lock()
		googleAuthSessions = map[string]*googleAuthSession{}
		googleAuthSessionsMu.Unlock()

		req := httptest.NewRequest("GET", "/auth/google/exchange?exchange_code=unknown-code", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Invalid or expired exchange code"}`, w.Body.String())
	})

	t.Run("CodeWithoutCompletedSession", func(t *testing.T) {
		user := createTestUser(t)

		googleAuthSessionsMu.Lock()
		googleAuthSessions = map[string]*googleAuthSession{}
		googleAuthSessions["web-session"] = &googleAuthSession{
			Mode:         googleAuthModeWeb,
			RedirectURL:  "http://localhost:5173/auth/google/callback",
			ExpiresAt:    time.Now().Add(5 * time.Minute),
			User:         &user,
			ExchangeCode: "incomplete-code",
		}
		googleAuthSessionsMu.Unlock()

		req := httptest.NewRequest("GET", "/auth/google/exchange?exchange_code=incomplete-code", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Invalid or expired exchange code"}`, w.Body.String())
	})
}

func TestGoogleCallbackRedirectsWithExchangeCodeForWeb(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	googleAuthSessionsMu.Lock()
	googleAuthSessions = map[string]*googleAuthSession{}
	googleAuthSessions["state-123"] = &googleAuthSession{
		Mode:        googleAuthModeWeb,
		RedirectURL: "http://localhost:5173/auth/google/callback",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		Token:       "jwt-token",
		User: &model.User{
			ID:    42,
			Name:  "Google User",
			Email: "google@example.com",
		},
	}
	googleAuthSessionsMu.Unlock()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("GET", "/auth/google/callback?code=oauth-code&state=state-123", nil)
	req.URL.RawQuery = "code=oauth-code&state=state-123"
	c.Request = req

	googleAuthSessionsMu.Lock()
	session := googleAuthSessions["state-123"]
	session.ExchangeCode = "exchange-code-generated"
	googleAuthSessionsMu.Unlock()

	redirectValues := url.Values{}
	redirectValues.Set("exchange_code", "exchange-code-generated")
	c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/auth/google/callback?"+redirectValues.Encode())

	assert.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	assert.Equal(
		t,
		"http://localhost:5173/auth/google/callback?exchange_code=exchange-code-generated",
		recorder.Header().Get("Location"),
	)
}
