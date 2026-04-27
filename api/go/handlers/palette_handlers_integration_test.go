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
	token, err := authpkg.GenerateJWTToken(user)
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
	if err := db.DB.Model(&model.Palette{}).Count(&count).Error; err != nil {
		t.Fatalf("count palettes: %v", err)
	}
	assert.Equal(t, int64(1), count)
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

func TestGetPalettesHandler_ReturnsSavedPalettes(t *testing.T) {
	setupTestDB(t)
	resetTestDB(t)

	user := createTestUser(t)
	if err := saveUserPalette(user.ID, "Saved", []Color{{Hex: "#112233"}}); err != nil {
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
	if err := saveUserPalette(user.ID, "To Delete", []Color{{Hex: "#FF0000"}}); err != nil {
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
	if err := saveUserPalette(user1.ID, "User1 Palette", []Color{{Hex: "#FF0000"}}); err != nil {
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
	if err := saveUserPalette(user.ID, "Palette 1", []Color{{Hex: "#FF0000"}}); err != nil {
		t.Fatalf("save palette1: %v", err)
	}
	if err := saveUserPalette(user.ID, "Palette 2", []Color{{Hex: "#00FF00"}}); err != nil {
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
	if err := saveUserPalette(owner.ID, "Owner Palette", []Color{{Hex: "#AA0000"}}); err != nil {
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
	if err := saveUserPalette(otherUser.ID, "Other Palette", []Color{{Hex: "#00AA00"}}); err != nil {
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
