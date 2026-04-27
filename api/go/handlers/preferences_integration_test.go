package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	authpkg "themesmith/auth"
	"themesmith/db"
	"themesmith/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

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
		JsonData: datatypes.JSON(`{"theme":"light"}`),
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
		JsonData: datatypes.JSON(`{"theme":"dark","language":"fr"}`),
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
