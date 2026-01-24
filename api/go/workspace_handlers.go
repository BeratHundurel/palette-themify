package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WorkspaceStateData struct {
	Colors             []Color    `json:"colors"`
	Selectors          []Selector `json:"selectors"`
	ActiveSelectorId   string     `json:"activeSelectorId"`
	Luminosity         float64    `json:"luminosity"`
	Nearest            int        `json:"nearest"`
	Power              int        `json:"power"`
	MaxDistance        float64    `json:"maxDistance"`
}

type Selector struct {
	ID        string     `json:"id"`
	Color     string     `json:"color"`
	Selected  bool       `json:"selected"`
	Selection *Selection `json:"selection,omitempty"`
}

type Selection struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type WorkspaceData struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	ImageData          string     `json:"imageData"`
	Colors             []Color    `json:"colors"`
	Selectors          []Selector `json:"selectors"`
	ActiveSelectorId   string     `json:"activeSelectorId"`
	Luminosity         float64    `json:"luminosity"`
	Nearest            int        `json:"nearest"`
	Power              int        `json:"power"`
	MaxDistance        float64    `json:"maxDistance"`
	ShareToken         *string    `json:"shareToken,omitempty"`
	CreatedAt          string     `json:"createdAt"`
}

type SaveWorkspaceRequest struct {
	Name               string     `json:"name" binding:"required"`
	ImageData          string     `json:"imageData" binding:"required"`
	Colors             []Color    `json:"colors"`
	Selectors          []Selector `json:"selectors"`
	ActiveSelectorId   string     `json:"activeSelectorId"`
	Luminosity         float64    `json:"luminosity"`
	Nearest            int        `json:"nearest"`
	Power              int        `json:"power"`
	MaxDistance        float64    `json:"maxDistance"`
}

type GetWorkspacesResponse struct {
	Workspaces []WorkspaceData `json:"workspaces"`
}

func saveWorkspaceHandler(c *gin.Context) {
	var req SaveWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authenticated, userID := isAuthenticated(c)

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save workspaces"})
		return
	}

	err := saveUserWorkspace(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save workspace"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Workspace saved successfully",
		"name":    req.Name,
	})
}

func getWorkspacesHandler(c *gin.Context) {
	authenticated, userID := isAuthenticated(c)

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to get workspaces"})
		return
	}

	workspaces, err := getUserWorkspaces(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
		return
	}

	c.JSON(http.StatusOK, GetWorkspacesResponse{Workspaces: workspaces})
}

func deleteWorkspaceHandler(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace ID is required"})
		return
	}

	authenticated, userID := isAuthenticated(c)

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to delete workspaces"})
		return
	}

	err := deleteUserWorkspace(userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace deleted successfully"})
}

func saveUserWorkspace(userID uint, req SaveWorkspaceRequest) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}

	stateData := WorkspaceStateData{
		Colors:             req.Colors,
		Selectors:          req.Selectors,
		ActiveSelectorId:   req.ActiveSelectorId,
		Luminosity:         req.Luminosity,
		Nearest:            req.Nearest,
		Power:              req.Power,
		MaxDistance:        req.MaxDistance,
	}

	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return err
	}

	dbWorkspace := Workspace{
		UserID:    &userID,
		Name:      req.Name,
		JsonData:  string(stateJSON),
		ImageData: req.ImageData,
	}

	return DB.Create(&dbWorkspace).Error
}

func getUserWorkspaces(userID uint) ([]WorkspaceData, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var dbWorkspaces []Workspace
	err := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dbWorkspaces).Error

	if err != nil {
		return nil, err
	}

	workspaces := make([]WorkspaceData, len(dbWorkspaces))
	for i, dbWorkspace := range dbWorkspaces {
		var state WorkspaceStateData
		if err := json.Unmarshal([]byte(dbWorkspace.JsonData), &state); err != nil {
			continue
		}

		workspaces[i] = WorkspaceData{
			ID:                 fmt.Sprintf("%d", dbWorkspace.ID),
			Name:               dbWorkspace.Name,
			ImageData:          dbWorkspace.ImageData,
			Colors:             state.Colors,
			Selectors:          state.Selectors,
			ActiveSelectorId:   state.ActiveSelectorId,
			Luminosity:         state.Luminosity,
			Nearest:            state.Nearest,
			Power:              state.Power,
			MaxDistance:        state.MaxDistance,
			ShareToken:         dbWorkspace.ShareToken,
			CreatedAt:          dbWorkspace.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		}
	}

	return workspaces, nil
}

func deleteUserWorkspace(userID uint, workspaceID string) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}

	var workspace Workspace
	if err := DB.Where("id = ? AND user_id = ?", workspaceID, userID).First(&workspace).Error; err != nil {
		return fmt.Errorf("workspace not found or unauthorized")
	}

	result := DB.Delete(&workspace)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func generateShareToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func shareWorkspaceHandler(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace ID is required"})
		return
	}

	authenticated, userID := isAuthenticated(c)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to share workspaces"})
		return
	}

	shareToken, err := createWorkspaceShareToken(userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shareToken": shareToken,
		"shareUrl":   fmt.Sprintf("http://localhost:5173/shared/%s", shareToken),
	})
}

func getSharedWorkspaceHandler(c *gin.Context) {
	shareToken := c.Param("token")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share token is required"})
		return
	}

	workspace, err := getWorkspaceByShareToken(shareToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shared workspace not found"})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func removeWorkspaceShareHandler(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workspace ID is required"})
		return
	}

	authenticated, userID := isAuthenticated(c)
	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	err := removeWorkspaceShareToken(userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Share link removed successfully"})
}

func createWorkspaceShareToken(userID uint, workspaceID string) (string, error) {
	if DB == nil {
		return "", fmt.Errorf("database not available")
	}

	var workspace Workspace
	if err := DB.Where("id = ? AND user_id = ?", workspaceID, userID).First(&workspace).Error; err != nil {
		return "", fmt.Errorf("workspace not found or unauthorized")
	}

	if workspace.ShareToken != nil {
		return *workspace.ShareToken, nil
	}

	shareToken, err := generateShareToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate share token")
	}

	workspace.ShareToken = &shareToken
	if err := DB.Save(&workspace).Error; err != nil {
		return "", fmt.Errorf("failed to save share token")
	}

	return shareToken, nil
}

func getWorkspaceByShareToken(shareToken string) (*WorkspaceData, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var dbWorkspace Workspace
	if err := DB.Where("share_token = ?", shareToken).First(&dbWorkspace).Error; err != nil {
		return nil, fmt.Errorf("workspace not found")
	}

	var state WorkspaceStateData
	if err := json.Unmarshal([]byte(dbWorkspace.JsonData), &state); err != nil {
		return nil, fmt.Errorf("failed to parse workspace data")
	}

	workspace := &WorkspaceData{
		ID:                 fmt.Sprintf("%d", dbWorkspace.ID),
		Name:               dbWorkspace.Name,
		ImageData:          dbWorkspace.ImageData,
		Colors:             state.Colors,
		Selectors:          state.Selectors,
		ActiveSelectorId:   state.ActiveSelectorId,
		Luminosity:         state.Luminosity,
		Nearest:            state.Nearest,
		Power:              state.Power,
		MaxDistance:        state.MaxDistance,
		ShareToken:         dbWorkspace.ShareToken,
		CreatedAt:          dbWorkspace.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	}

	return workspace, nil
}

func removeWorkspaceShareToken(userID uint, workspaceID string) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}

	var workspace Workspace
	if err := DB.Where("id = ? AND user_id = ?", workspaceID, userID).First(&workspace).Error; err != nil {
		return fmt.Errorf("workspace not found or unauthorized")
	}

	workspace.ShareToken = nil
	if err := DB.Save(&workspace).Error; err != nil {
		return fmt.Errorf("failed to remove share token")
	}

	return nil
}
