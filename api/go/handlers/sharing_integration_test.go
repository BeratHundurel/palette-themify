package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	authpkg "themesmith/auth"
	"themesmith/db"
	"themesmith/model"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestShareThemeHandler_SetsSharedFlags(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	saved, _, err := createTestTheme(user.ID, "Theme", "vscode", "sig-share", `{"name":"Theme"}`)
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
	saved, _, err := createTestTheme(user.ID, "Theme", "vscode", "sig-unshare", `{"name":"Theme"}`)
	if err != nil {
		t.Fatalf("save theme: %v", err)
	}

	now := time.Now()

	err = db.DB.Model(&model.Theme{}).Where("id = ? AND user_id = ?", saved.ID, user.ID).Updates(model.Theme{
		IsShared: true,
		SharedAt: &now,
	}).Error

	if err != nil {
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
}

func TestGetSharedItemsHandler_ReturnsSharedPalettesAndThemes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "Shared Palette", datatypes.JSON(`[{"hex": "#112233"}, {"hex": "#445566"}]`)); err != nil {
		t.Fatalf("save palette: %v", err)
	}

	var savedPalette model.Palette
	if err := db.DB.Where("user_id = ?", user.ID).First(&savedPalette).Error; err != nil {
		t.Fatalf("load palette: %v", err)
	}

	if _, err := setPaletteShared(user.ID, fmt.Sprintf("%d", savedPalette.ID), true); err != nil {
		t.Fatalf("share palette: %v", err)
	}

	_, _, err := createTestTheme(
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

	now := time.Now().UTC()
	err = db.DB.Model(&model.Theme{}).Where("id = ? AND user_id = ?", savedTheme.ID, user.ID).Updates(model.Theme{
		IsShared: true,
		SharedAt: &now,
	}).Error

	if err != nil {
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
	assert.Equal(t, "Shared Theme", resp.Items[1].Name)
}
