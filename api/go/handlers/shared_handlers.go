package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"themesmith/db"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type SharedItemSort string

const (
	SharedItemSortNewest SharedItemSort = "newest"
	SharedItemSortOldest SharedItemSort = "oldest"
	SharedItemSortName   SharedItemSort = "name"
)

type sharedItemView struct {
	Kind       string         `json:"kind" gorm:"column:kind"`
	ItemID     uint           `json:"itemId" gorm:"column:item_id"`
	Name       string         `json:"name" gorm:"column:name"`
	JsonData   datatypes.JSON `json:"jsonData" gorm:"column:json_data"`
	SharedAt   time.Time      `json:"sharedAt" gorm:"column:shared_at"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"column:created_at"`
	EditorType *string        `json:"editorType,omitempty" gorm:"column:editor_type"`
	Signature  *string        `json:"signature,omitempty" gorm:"column:signature"`
}

type SharedItemsResponse struct {
	Items []sharedItemView `json:"items"`
}

func GetSharedItemsHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	sort := parseSharedSort(c.Query("sort"))
	limit := parsePositiveInt(c.Query("limit"), 100)

	items, err := listSharedItems(query, sort, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared items"})
		return
	}

	c.JSON(http.StatusOK, SharedItemsResponse{Items: items})
}

func listSharedItems(query string, sort SharedItemSort, limit int) ([]sharedItemView, error) {
	paletteFilter := "WHERE is_shared = true AND shared_at IS NOT NULL"
	themeFilter := "WHERE is_shared = true AND shared_at IS NOT NULL"
	args := make([]any, 0, 3)
	if query != "" {
		paletteFilter += " AND name ILIKE ?"
		args = append(args, "%"+query+"%")
		themeFilter += " AND name ILIKE ?"
		args = append(args, "%"+query+"%")
	}

	orderBy := "shared_at DESC, lower(name) ASC"
	switch sort {
	case SharedItemSortOldest:
		orderBy = "shared_at ASC, lower(name) ASC"
	case SharedItemSortName:
		orderBy = "lower(name) ASC, shared_at DESC"
	}

	querySQL := fmt.Sprintf(`
		SELECT * FROM (
			SELECT 'palette' AS kind,
				id AS item_id,
				name,
				json_data,
				shared_at,
				created_at,
				NULL::text AS editor_type,
				NULL::text AS signature
			FROM palettes
			%s
			UNION ALL
			SELECT 'theme' AS kind,
				id AS item_id,
				name,
				json_data,
				shared_at,
				created_at,
				editor_type,
				signature
			FROM themes
			%s
		) AS shared_items
		ORDER BY %s
		LIMIT ?`, paletteFilter, themeFilter, orderBy)

	args = append(args, limit)

	var rows []sharedItemView
	if err := db.DB.Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func parseSharedSort(raw string) SharedItemSort {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(SharedItemSortOldest):
		return SharedItemSortOldest
	case string(SharedItemSortName):
		return SharedItemSortName
	default:
		return SharedItemSortNewest
	}
}

func parsePositiveInt(raw string, fallback int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	if v > 500 {
		return 500
	}
	return v
}
