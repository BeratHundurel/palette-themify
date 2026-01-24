package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PaletteData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Palette   []Color   `json:"palette"`
	CreatedAt time.Time `json:"createdAt"`
	IsSystem  bool      `json:"isSystem"`
}

type SavePaletteRequest struct {
	Name    string  `json:"name" binding:"required"`
	Palette []Color `json:"palette" binding:"required"`
}

type GetPalettesResponse struct {
	Palettes []PaletteData `json:"palettes"`
}

func isAuthenticated(c *gin.Context) (bool, uint) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false, 0
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		return false, 0
	}

	claims, err := validateJWTToken(bearerToken[1])
	if err != nil {
		return false, 0
	}

	return true, claims.UserID
}

func savePaletteHandler(c *gin.Context) {
	var req SavePaletteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authenticated, userID := isAuthenticated(c)

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save palettes"})
		return
	}

	err := saveUserPalette(userID, req.Name, req.Palette)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save palette"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Palette saved successfully",
		"name":    req.Name,
	})
}

func getPalettesHandler(c *gin.Context) {
	authenticated, userID := isAuthenticated(c)

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to get palettes"})
		return
	}

	palettes, err := getUserPalettes(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch palettes"})
		return
	}

	c.JSON(http.StatusOK, GetPalettesResponse{Palettes: palettes})
}

func deletePaletteHandler(c *gin.Context) {
	paletteID := c.Param("id")
	if paletteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette ID is required"})
		return
	}

	authenticated, userID := isAuthenticated(c)

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to delete palettes"})
		return
	}

	err := deleteUserPalette(userID, paletteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Palette deleted successfully"})
}

func saveUserPalette(userID uint, name string, palette []Color) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}

	paletteJSON, err := json.Marshal(palette)
	if err != nil {
		return err
	}

	dbPalette := Palette{
		UserID:   &userID,
		JsonData: string(paletteJSON),
		Name:     name,
	}

	return DB.Create(&dbPalette).Error
}

func getUserPalettes(userID uint) ([]PaletteData, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var dbPalettes []Palette
	err := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dbPalettes).Error
	if err != nil {
		return nil, err
	}

	palettes := make([]PaletteData, len(dbPalettes))
	for i, dbPalette := range dbPalettes {
		var colors []Color
		if err := json.Unmarshal([]byte(dbPalette.JsonData), &colors); err != nil {
			continue
		}

		palettes[i] = PaletteData{
			ID:        fmt.Sprintf("%d", dbPalette.ID),
			Name:      dbPalette.Name,
			Palette:   colors,
			CreatedAt: dbPalette.CreatedAt,
			IsSystem:  dbPalette.IsSystem,
		}
	}

	return palettes, nil
}

func deleteUserPalette(userID uint, paletteID string) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}

	var palette Palette
	if err := DB.Where("id = ? AND user_id = ?", paletteID, userID).First(&palette).Error; err != nil {
		return fmt.Errorf("palette not found or unauthorized")
	}

	if palette.IsSystem {
		return fmt.Errorf("cannot delete system palettes")
	}

	result := DB.Delete(&palette)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
