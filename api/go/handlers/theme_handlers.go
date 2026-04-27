package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"themesmith/auth"
	"themesmith/db"
	"themesmith/model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

type ThemeDTO struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	EditorType  string         `json:"editorType"`
	Signature   string         `json:"signature"`
	ThemeResult datatypes.JSON `json:"themeResult"`
	CreatedAt   time.Time      `json:"createdAt"`
	IsShared    bool           `json:"isShared"`
	SharedAt    *time.Time     `json:"sharedAt"`
}

type SaveThemeDTO struct {
	Name        string         `json:"name"`
	EditorType  string         `json:"editorType"`
	Signature   string         `json:"signature"`
	ThemeResult datatypes.JSON `json:"themeResult"`
	CreatedAt   time.Time      `json:"createdAt"`
	IsShared    bool           `json:"isShared"`
	SharedAt    *time.Time     `json:"sharedAt"`
}

type ThemeRequest struct {
	Theme ThemeDTO `json:"theme"`
}

type ThemeResponse struct {
	Theme ThemeDTO `json:"theme"`
}

type ThemesRequest struct {
	Themes []ThemeDTO `json:"themes"`
}

type ThemesSaveRequest struct {
	Themes []SaveThemeDTO `json:"themes"`
}

type ThemesResponse struct {
	Themes []ThemeDTO `json:"themes"`
}

type UpdateThemeRequest struct {
	ThemeId string   `json:"themeId"`
	Theme   ThemeDTO `json:"theme"`
}

type DeleteThemesRequest struct {
	IDs []string `json:"ids"`
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

	var themeSaveDTO SaveThemeDTO
	if err := c.ShouldBindJSON(&themeSaveDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme payload"})
		return
	}

	themeModel := modelFromSaveThemeDTO(themeSaveDTO, &userID)
	err = db.DB.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "editor_type"},
				{Name: "signature"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"json_data",
				"updated_at",
			}),
		}).
		Create(&themeModel).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	themeDTO := themeDTOFromModel(themeModel)

	c.JSON(http.StatusCreated, ThemeResponse{Theme: themeDTO})
}

func SaveThemesBatchHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save themes"})
		return
	}

	var req ThemesSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if len(req.Themes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Themes are required"})
		return
	}

	themeModels := modelsFromSaveThemeDTOs(req.Themes, &userID)

	seen := map[string]bool{}
	unique := make([]model.Theme, 0, len(themeModels))

	for _, t := range themeModels {
		key := fmt.Sprintf("%v|%s|%s", *t.UserID, t.EditorType, t.Signature)

		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, t)
	}

	err = db.DB.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "editor_type"},
				{Name: "signature"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"json_data",
				"updated_at",
			}),
		}).
		Create(&unique).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	themeDTOs := themeDTOsFromModels(unique)

	c.JSON(http.StatusOK, ThemesResponse{Themes: themeDTOs})
}

func UpdateThemeHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to update themes"})
		return
	}

	themeID := c.Param("id")
	if themeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme ID is required"})
		return
	}

	var themeDTO ThemeDTO
	if err := c.ShouldBindJSON(&themeDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	themeModel := modelFromThemeDTO(themeDTO, &userID)
	result := db.DB.Where("id = ? AND user_id = ?", themeID, userID).Updates(&themeModel)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theme"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Theme not found"})
		return
	}

	themeDTO = themeDTOFromModel(themeModel)

	c.JSON(http.StatusOK, ThemeResponse{Theme: themeDTO})
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

	var dbThemes []model.Theme
	if err := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&dbThemes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch themes"})
		return
	}

	themeDTOs := themeDTOsFromModels(dbThemes)
	c.JSON(http.StatusOK, ThemesResponse{Themes: themeDTOs})
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

	result := db.DB.Where("id = ? AND user_id = ?", themeID, userID).Delete(&model.Theme{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Theme not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteThemesBatchHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	var req DeleteThemesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme IDs are required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to delete themes"})
		return
	}

	if err := db.DB.Where("id IN ? AND user_id = ?", req.IDs, userID).Delete(&model.Theme{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func ShareThemeHandler(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to share themes"})
		return
	}

	dto, err := setThemeShared(themeID, userID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto)
}

func UnshareThemeHandler(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to unshare themes"})
		return
	}

	dto, err := setThemeShared(themeID, userID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto)
}

func setThemeShared(themeID string, userID uint, isShared bool) (ThemeDTO, error) {
	var theme model.Theme
	err := db.DB.Model(&model.Theme{}).
		Where("id = ? AND user_id = ?", themeID, userID).
		Clauses(clause.Returning{}).
		Updates(map[string]any{
			"is_shared": isShared,
			"shared_at": time.Now().UTC(),
		}).
		Scan(&theme).Error

	if err != nil {
		return ThemeDTO{}, err
	}

	return themeDTOFromModel(theme), nil
}

func normalizeThemeSignature(signature string) string {
	if len(signature) <= 128 {
		return signature
	}

	hash := sha256.Sum256([]byte(signature))
	return hex.EncodeToString(hash[:])
}

func themeDTOFromModel(theme model.Theme) ThemeDTO {
	return ThemeDTO{
		ID:          theme.ID,
		Name:        theme.Name,
		EditorType:  theme.EditorType,
		Signature:   theme.Signature,
		ThemeResult: theme.JsonData,
		IsShared:    theme.IsShared,
		SharedAt:    theme.SharedAt,
		CreatedAt:   theme.CreatedAt,
	}
}

func themeDTOsFromModels(themes []model.Theme) []ThemeDTO {
	result := make([]ThemeDTO, len(themes))
	for i, theme := range themes {
		result[i] = themeDTOFromModel(theme)
	}
	return result
}

func modelFromThemeDTO(dto ThemeDTO, userID *uint) model.Theme {
	return model.Theme{
		ID:         dto.ID,
		UserID:     userID,
		Name:       dto.Name,
		EditorType: dto.EditorType,
		Signature:  normalizeThemeSignature(dto.Signature),
		JsonData:   dto.ThemeResult,
		IsShared:   dto.IsShared,
		SharedAt:   dto.SharedAt,
		CreatedAt:  dto.CreatedAt,
		UpdatedAt:  time.Now().UTC(),
	}
}

func modelsFromThemeDTOs(dtos []ThemeDTO, userID *uint) []model.Theme {
	result := make([]model.Theme, len(dtos))
	for i, dto := range dtos {
		result[i] = modelFromThemeDTO(dto, userID)
	}
	return result
}

func modelFromSaveThemeDTO(dto SaveThemeDTO, userID *uint) model.Theme {
	return model.Theme{
		UserID:     userID,
		Name:       dto.Name,
		EditorType: dto.EditorType,
		Signature:  normalizeThemeSignature(dto.Signature),
		JsonData:   dto.ThemeResult,
		IsShared:   dto.IsShared,
		SharedAt:   dto.SharedAt,
		CreatedAt:  dto.CreatedAt,
		UpdatedAt:  time.Now().UTC(),
	}
}

func modelsFromSaveThemeDTOs(dtos []SaveThemeDTO, userID *uint) []model.Theme {
	result := make([]model.Theme, len(dtos))
	for i, dto := range dtos {
		result[i] = modelFromSaveThemeDTO(dto, userID)
	}
	return result
}
