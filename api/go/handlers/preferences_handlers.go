package handlers

import (
	"encoding/json"
	"errors"
	"image-to-palette/auth"
	"image-to-palette/db"
	"image-to-palette/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PreferencesResponse struct {
	Preferences json.RawMessage `json:"preferences"`
}

func GetPreferencesHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := auth.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var prefs model.UserPreferences
	if err := db.DB.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"preferences": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preferences"})
		return
	}

	c.JSON(http.StatusOK, PreferencesResponse{Preferences: json.RawMessage(prefs.JsonData)})
}

func SavePreferencesHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := auth.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var payload json.RawMessage
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var prefs model.UserPreferences
	if err := db.DB.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			return
		}

		prefs = model.UserPreferences{
			UserID:   userID,
			JsonData: string(payload),
		}
		if err := db.DB.Create(&prefs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			return
		}
	} else {
		prefs.JsonData = string(payload)
		if err := db.DB.Save(&prefs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			return
		}
	}

	c.JSON(http.StatusOK, PreferencesResponse{Preferences: json.RawMessage(prefs.JsonData)})
}
