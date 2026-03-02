package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PreferencesResponse struct {
	Preferences json.RawMessage `json:"preferences"`
}

func getPreferencesHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var prefs UserPreferences
	if err := DB.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"preferences": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preferences"})
		return
	}

	c.JSON(http.StatusOK, PreferencesResponse{Preferences: json.RawMessage(prefs.JsonData)})
}

func savePreferencesHandler(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var payload json.RawMessage
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var prefs UserPreferences
	if err := DB.Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			return
		}

		prefs = UserPreferences{
			UserID:   userID,
			JsonData: string(payload),
		}
		if err := DB.Create(&prefs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			return
		}
	} else {
		prefs.JsonData = string(payload)
		if err := DB.Save(&prefs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			return
		}
	}

	c.JSON(http.StatusOK, PreferencesResponse{Preferences: json.RawMessage(prefs.JsonData)})
}
