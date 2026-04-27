package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	authpkg "themesmith/auth"
	"themesmith/db"
	"themesmith/model"

	"github.com/stretchr/testify/assert"
)

func TestSaveThemesBatchHandler_AuthRequired(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	router := setupThemeRouter()

	body, err := json.Marshal(map[string]any{"themes": []any{buildThemePayload("Theme", "vscode", "sig-1")}})
	if err != nil {
		t.Fatalf("marshal themes: %v", err)
	}

	req := httptest.NewRequest("POST", "/themes/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSaveThemesBatchHandler_SavesAndDedupes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	token, err := authpkg.GenerateJWTToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := setupThemeRouter()

	body, err := json.Marshal(map[string]any{
		"themes": []any{
			buildThemePayload("Theme One", "vscode", "sig-batch-1"),
			buildThemePayload("Theme One", "vscode", "sig-batch-1"),
			buildThemePayload("Theme Two", "zed", "sig-batch-2"),
		},
	})
	if err != nil {
		t.Fatalf("marshal themes: %v", err)
	}

	req := httptest.NewRequest("POST", "/themes/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ThemesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assert.Len(t, resp.Themes, 2)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(2), count)
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
	assert.Equal(t, http.StatusCreated, second.Code)

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
	assert.Equal(t, http.StatusCreated, second.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(1), count)
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

func TestGetThemesHandler_ReturnsThemes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	_, _, err := createTestTheme(user.ID, "Theme", "vscode", "sig-1", `{"name":"Theme"}`)
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

	var resp ThemesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assert.Len(t, resp.Themes, 1)
}

func TestUpdateThemeHandler_UpdatesPayload(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := createTestTheme(user.ID, "Theme", "vscode", "sig-1", `{"name":"Theme"}`)
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

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteThemeHandler_RemovesTheme(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := createTestTheme(user.ID, "Theme", "vscode", "sig-1", `{"name":"Theme"}`)
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

	assert.Equal(t, http.StatusNoContent, w.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(0), count)
}

func TestDeleteThemeHandler_CannotDeleteOtherUsersTheme(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user1 := createTestUser(t)
	theme1, _, err := createTestTheme(user1.ID, "User1 Theme", "vscode", "sig-1", `{"name":"Theme"}`)
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

	assert.Equal(t, http.StatusNotFound, w.Code)

	var count int64
	if err := db.DB.Model(&model.Theme{}).Count(&count).Error; err != nil {
		t.Fatalf("count themes: %v", err)
	}
	assert.Equal(t, int64(1), count, "Theme should not be deleted")
}

func TestDeleteThemesBatchHandler_RemovesOnlyRequestedThemes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	theme1, _, err := createTestTheme(user.ID, "Theme 1", "vscode", "sig-1", `{"name":"Theme 1"}`)
	if err != nil {
		t.Fatalf("save theme1: %v", err)
	}
	theme2, _, err := createTestTheme(user.ID, "Theme 2", "vscode", "sig-2", `{"name":"Theme 2"}`)
	if err != nil {
		t.Fatalf("save theme2: %v", err)
	}
	_, _, err = createTestTheme(user.ID, "Theme 3", "vscode", "sig-3", `{"name":"Theme 3"}`)
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

	assert.Equal(t, http.StatusNoContent, w.Code)

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
	ownerTheme, _, err := createTestTheme(owner.ID, "Owner Theme", "vscode", "owner-sig", `{"name":"Owner Theme"}`)
	if err != nil {
		t.Fatalf("save owner theme: %v", err)
	}

	otherUser := model.User{Name: "Other User", Email: "other@example.com", PasswordHash: "hash"}
	if err := db.DB.Create(&otherUser).Error; err != nil {
		t.Fatalf("create other user: %v", err)
	}
	otherTheme, _, err := createTestTheme(otherUser.ID, "Other Theme", "vscode", "other-sig", `{"name":"Other Theme"}`)
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

	assert.Equal(t, http.StatusNoContent, w.Code)

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
