package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image-to-palette/auth"
	"image-to-palette/db"
	"image-to-palette/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GetThemesResponse struct {
	Themes []json.RawMessage `json:"themes"`
}

type themePayload struct {
	Name       string
	EditorType string
	Signature  string
}

func SaveThemeHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save themes"})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme payload"})
		return
	}

	payload, info, err := parseThemePayload(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	theme, created, err := saveUserTheme(userID, info.Name, info.EditorType, info.Signature, string(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save theme"})
		return
	}

	responseTheme, err := buildThemeResponse(theme, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build theme response"})
		return
	}

	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}

	c.JSON(status, gin.H{"message": "Theme saved successfully", "theme": responseTheme})
}

func UpdateThemeHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	themeID := c.Param("id")
	if themeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme ID is required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to update themes"})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme payload"})
		return
	}

	payload, info, err := parseThemePayload(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	theme, err := updateUserTheme(userID, themeID, info.Name, info.EditorType, info.Signature, string(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responseTheme, err := buildThemeResponse(theme, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build theme response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme updated successfully", "theme": responseTheme})
}

func GetThemesHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to get themes"})
		return
	}

	themes, err := getUserThemes(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch themes"})
		return
	}

	c.JSON(http.StatusOK, GetThemesResponse{Themes: themes})
}

func DeleteThemeHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	themeID := c.Param("id")
	if themeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme ID is required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to delete themes"})
		return
	}

	if err := deleteUserTheme(userID, themeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme deleted successfully"})
}

func saveUserTheme(userID uint, name string, editorType string, signature string, jsonData string) (model.Theme, bool, error) {
	if db.DB == nil {
		return model.Theme{}, false, fmt.Errorf("database not available")
	}

	var theme model.Theme
	err := db.DB.Where("user_id = ? AND editor_type = ? AND signature = ?", userID, editorType, signature).First(&theme).Error
	if err == nil {
		theme.Name = name
		theme.JsonData = jsonData
		theme.UpdatedAt = time.Now().UTC()
		if err := db.DB.Save(&theme).Error; err != nil {
			return model.Theme{}, false, err
		}
		return theme, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Theme{}, false, err
	}

	newTheme := model.Theme{
		UserID:     &userID,
		Name:       name,
		EditorType: editorType,
		Signature:  signature,
		JsonData:   jsonData,
	}

	if err := db.DB.Create(&newTheme).Error; err != nil {
		return model.Theme{}, false, err
	}

	return newTheme, true, nil
}

func updateUserTheme(userID uint, themeID string, name string, editorType string, signature string, jsonData string) (model.Theme, error) {
	if db.DB == nil {
		return model.Theme{}, fmt.Errorf("database not available")
	}

	var theme model.Theme
	if err := db.DB.Where("id = ? AND user_id = ?", themeID, userID).First(&theme).Error; err != nil {
		return model.Theme{}, fmt.Errorf("theme not found or unauthorized")
	}

	theme.Name = name
	theme.EditorType = editorType
	theme.Signature = signature
	theme.JsonData = jsonData
	theme.UpdatedAt = time.Now().UTC()

	if err := db.DB.Save(&theme).Error; err != nil {
		return model.Theme{}, err
	}

	return theme, nil
}

func getUserThemes(userID uint) ([]json.RawMessage, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var dbThemes []model.Theme
	if err := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&dbThemes).Error; err != nil {
		return nil, err
	}

	themes := make([]json.RawMessage, 0, len(dbThemes))
	for _, dbTheme := range dbThemes {
		payload, err := decodeThemePayload(dbTheme.JsonData)
		if err != nil {
			continue
		}
		payload["id"] = fmt.Sprintf("%d", dbTheme.ID)
		payload["name"] = dbTheme.Name
		payload["editorType"] = dbTheme.EditorType
		payload["signature"] = dbTheme.Signature
		encoded, err := json.Marshal(payload)
		if err != nil {
			continue
		}
		themes = append(themes, json.RawMessage(encoded))
	}

	return themes, nil
}

func deleteUserTheme(userID uint, themeID string) error {
	if db.DB == nil {
		return fmt.Errorf("database not available")
	}

	var theme model.Theme
	if err := db.DB.Where("id = ? AND user_id = ?", themeID, userID).First(&theme).Error; err != nil {
		return fmt.Errorf("theme not found or unauthorized")
	}

	return db.DB.Delete(&theme).Error
}

func parseThemePayload(body []byte) (map[string]any, themePayload, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, themePayload{}, fmt.Errorf("invalid theme payload")
	}

	name, _ := payload["name"].(string)
	editorType, _ := payload["editorType"].(string)
	signature, _ := payload["signature"].(string)

	if name == "" {
		return nil, themePayload{}, fmt.Errorf("theme name is required")
	}
	if editorType == "" {
		return nil, themePayload{}, fmt.Errorf("theme editor type is required")
	}
	if signature == "" {
		return nil, themePayload{}, fmt.Errorf("theme signature is required")
	}

	signature = normalizeThemeSignature(signature)

	info := themePayload{
		Name:       name,
		EditorType: editorType,
		Signature:  signature,
	}

	return payload, info, nil
}

func normalizeThemeSignature(signature string) string {
	if len(signature) <= 128 {
		return signature
	}

	hash := sha256.Sum256([]byte(signature))
	return hex.EncodeToString(hash[:])
}

func decodeThemePayload(raw string) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func buildThemeResponse(theme model.Theme, payload map[string]any) (map[string]any, error) {
	if payload == nil {
		decoded, err := decodeThemePayload(theme.JsonData)
		if err != nil {
			return nil, err
		}
		payload = decoded
	}

	payload["id"] = fmt.Sprintf("%d", theme.ID)
	payload["name"] = theme.Name
	payload["editorType"] = theme.EditorType
	payload["signature"] = theme.Signature

	return payload, nil
}
