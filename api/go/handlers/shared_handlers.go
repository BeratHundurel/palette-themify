package handlers

import (
	"encoding/json"
	"fmt"
	"image-to-palette/db"
	"image-to-palette/model"
	"net/http"
	gsort "sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SharedItemKind string

const (
	SharedItemKindTheme   SharedItemKind = "theme"
	SharedItemKindPalette SharedItemKind = "palette"
)

type SharedItemSort string

const (
	SharedItemSortNewest SharedItemSort = "newest"
	SharedItemSortOldest SharedItemSort = "oldest"
	SharedItemSortName   SharedItemSort = "name"
)

type SharedItem struct {
	ID         string         `json:"id"`
	Kind       SharedItemKind `json:"kind"`
	Name       string         `json:"name"`
	Palette    []model.Color  `json:"palette"`
	SharedAt   time.Time      `json:"sharedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	EditorType string         `json:"editorType,omitempty"`
	Theme      any            `json:"theme,omitempty"`
}

type SharedItemsResponse struct {
	Items []SharedItem `json:"items"`
}

func GetSharedItemsHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	sort := parseSharedSort(c.Query("sort"))
	limit := parsePositiveInt(c.Query("limit"), 100)

	palettes, err := listSharedPalettes(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared palettes"})
		return
	}

	themes, err := listSharedThemes(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared themes"})
		return
	}

	items := append(palettes, themes...)
	sortSharedItems(items, sort)

	if len(items) > limit {
		items = items[:limit]
	}

	c.JSON(http.StatusOK, SharedItemsResponse{Items: items})
}

func listSharedPalettes(query string, limit int) ([]SharedItem, error) {
	q := db.DB.Where("is_shared = ?", true)
	if query != "" {
		q = q.Where("name ILIKE ?", "%"+query+"%")
	}

	var rows []model.Palette
	if err := q.Order("shared_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]SharedItem, 0, len(rows))
	for _, row := range rows {
		if row.SharedAt == nil {
			continue
		}
		var colors []model.Color
		if err := json.Unmarshal([]byte(row.JsonData), &colors); err != nil {
			continue
		}

		items = append(items, SharedItem{
			ID:        fmt.Sprintf("palette:%d", row.ID),
			Kind:      SharedItemKindPalette,
			Name:      row.Name,
			Palette:   colors,
			SharedAt:  *row.SharedAt,
			CreatedAt: row.CreatedAt,
		})
	}

	return items, nil
}

func listSharedThemes(query string, limit int) ([]SharedItem, error) {
	q := db.DB.Where("is_shared = ?", true)
	if query != "" {
		q = q.Where("name ILIKE ?", "%"+query+"%")
	}

	var rows []model.Theme
	if err := q.Order("shared_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]SharedItem, 0, len(rows))
	for _, row := range rows {
		if row.SharedAt == nil {
			continue
		}

		payload, err := decodeThemePayload(row.JsonData)
		if err != nil {
			continue
		}

		palette := extractPaletteFromThemePayload(payload)
		itemTheme := extractThemeObject(payload)

		items = append(items, SharedItem{
			ID:         fmt.Sprintf("theme:%d", row.ID),
			Kind:       SharedItemKindTheme,
			Name:       row.Name,
			Palette:    palette,
			SharedAt:   *row.SharedAt,
			CreatedAt:  row.CreatedAt,
			EditorType: row.EditorType,
			Theme:      itemTheme,
		})
	}

	return items, nil
}

func extractPaletteFromThemePayload(payload map[string]any) []model.Color {
	themeResult, ok := payload["themeResult"].(map[string]any)
	if !ok {
		return nil
	}

	colorsRaw, ok := themeResult["colors"].([]any)
	if !ok {
		return nil
	}

	colors := make([]model.Color, 0, len(colorsRaw))
	for _, raw := range colorsRaw {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		hex, _ := obj["hex"].(string)
		if hex == "" {
			continue
		}
		colors = append(colors, model.Color{Hex: hex})
	}

	return colors
}

func extractThemeObject(payload map[string]any) any {
	themeResult, ok := payload["themeResult"].(map[string]any)
	if !ok {
		return payload["theme"]
	}

	if theme, exists := themeResult["theme"]; exists {
		return theme
	}

	return payload["theme"]
}

func sortSharedItems(items []SharedItem, sortBy SharedItemSort) {
	switch sortBy {
	case SharedItemSortOldest:
		gsort.Slice(items, func(i, j int) bool {
			a := items[i]
			b := items[j]
			if a.SharedAt.Equal(b.SharedAt) {
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
			return a.SharedAt.Before(b.SharedAt)
		})
	case SharedItemSortName:
		gsort.Slice(items, func(i, j int) bool {
			a := items[i]
			b := items[j]
			nameA := strings.ToLower(strings.TrimSpace(a.Name))
			nameB := strings.ToLower(strings.TrimSpace(b.Name))
			if nameA == nameB {
				return a.SharedAt.After(b.SharedAt)
			}
			return nameA < nameB
		})
	default:
		gsort.Slice(items, func(i, j int) bool {
			a := items[i]
			b := items[j]
			if a.SharedAt.Equal(b.SharedAt) {
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
			return a.SharedAt.After(b.SharedAt)
		})
	}
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
