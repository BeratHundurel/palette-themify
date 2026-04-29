package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"themesmith/auth"
	"themesmith/db"
	"themesmith/model"
	"themesmith/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

type Color struct {
	Hex string `json:"hex"`
}

type PaletteDTO struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	Palette   datatypes.JSON `json:"palette"`
	CreatedAt time.Time      `json:"createdAt"`
	IsSystem  bool           `json:"isSystem"`
	IsShared  bool           `json:"isShared"`
	SharedAt  *time.Time     `json:"sharedAt"`
}

type SavePaletteRequest struct {
	Name    string         `json:"name" binding:"required"`
	Palette datatypes.JSON `json:"palette" binding:"required"`
}

type SavePalettesBatchRequest struct {
	Palettes []SavePaletteRequest `json:"palettes" binding:"required"`
}

type DeletePalettesRequest struct {
	IDs []string `json:"ids"`
}

type GetPalettesResponse struct {
	Palettes []PaletteDTO `json:"palettes"`
}

func SavePaletteHandler(c *gin.Context) {
	var req SavePaletteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save palettes"})
		return
	}

	err = saveUserPalette(userID, req.Name, req.Palette)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save palette"})
		return
	}

	c.Status(http.StatusCreated)
}

func SavePalettesBatchHandler(c *gin.Context) {
	var req SavePalettesBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Palettes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palettes are required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to save palettes"})
		return
	}

	if err := saveUserPalettesBatch(userID, req.Palettes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save palettes"})
		return
	}

	c.Status(http.StatusCreated)
}

func GetPalettesHandler(c *gin.Context) {
	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
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

func DeletePaletteHandler(c *gin.Context) {
	paletteID := c.Param("id")
	if paletteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette ID is required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to delete palettes"})
		return
	}

	err = deleteUserPalette(userID, paletteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func DeletePalettesBatchHandler(c *gin.Context) {
	var req DeletePalettesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette IDs are required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to delete palettes"})
		return
	}

	err = deleteUserPalettes(userID, req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func SharePaletteHandler(c *gin.Context) {
	paletteID := c.Param("id")
	if paletteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette ID is required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to share palettes"})
		return
	}

	palette, err := setPaletteShared(userID, paletteID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, palette)
}

func UnsharePaletteHandler(c *gin.Context) {
	paletteID := c.Param("id")
	if paletteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette ID is required"})
		return
	}

	userID, err := auth.GetUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required to unshare palettes"})
		return
	}

	palette, err := setPaletteShared(userID, paletteID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, palette)
}

func saveUserPalette(userID uint, name string, palette datatypes.JSON) error {
	if db.DB == nil {
		return fmt.Errorf("database not available")
	}

	dbPalette := model.Palette{
		UserID:   &userID,
		JsonData: palette,
		Name:     name,
	}

	return db.DB.Create(&dbPalette).Error
}

func saveUserPalettesBatch(userID uint, palettes []SavePaletteRequest) error {
	if db.DB == nil {
		return fmt.Errorf("database not available")
	}

	dbPalettes := make([]model.Palette, 0, len(palettes))
	for _, item := range palettes {
		paletteJSON, err := json.Marshal(item.Palette)
		if err != nil {
			return err
		}

		dbPalettes = append(dbPalettes, model.Palette{
			UserID:   &userID,
			JsonData: datatypes.JSON(paletteJSON),
			Name:     item.Name,
		})
	}

	return db.DB.Create(&dbPalettes).Error
}

func getUserPalettes(userID uint) ([]PaletteDTO, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var dbPalettes []model.Palette
	err := db.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dbPalettes).Error

	if err != nil {
		return nil, err
	}

	palettes := make([]PaletteDTO, len(dbPalettes))
	for i, dbPalette := range dbPalettes {

		palettes[i] = PaletteDTO{
			ID:        dbPalette.ID,
			Name:      dbPalette.Name,
			Palette:   dbPalette.JsonData,
			CreatedAt: dbPalette.CreatedAt,
			IsSystem:  dbPalette.IsSystem,
			IsShared:  dbPalette.IsShared,
			SharedAt:  dbPalette.SharedAt,
		}
	}

	return palettes, nil
}

func deleteUserPalette(userID uint, paletteID string) error {
	if db.DB == nil {
		return fmt.Errorf("database not available")
	}

	result := db.DB.
		Where("id = ? AND user_id = ? AND is_system = ?", paletteID, userID, false).
		Delete(&model.Palette{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete palette: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("palette not found, unauthorized, or system palette")
	}

	return nil
}

func deleteUserPalettes(userID uint, paletteIDs []string) error {
	if db.DB == nil {
		return fmt.Errorf("database not available")
	}

	result := db.DB.Where("id IN ? AND user_id = ? AND is_system = ?", paletteIDs, userID, false).Delete(&model.Palette{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func setPaletteShared(userID uint, paletteID string, shared bool) (PaletteDTO, error) {
	if db.DB == nil {
		return PaletteDTO{}, fmt.Errorf("database not available")
	}

	now := time.Now().UTC()
	var palette model.Palette
	err := db.DB.Model(&model.Palette{}).
		Where("id = ? AND user_id = ?", paletteID, userID).
		Clauses(clause.Returning{}).
		Updates(map[string]any{
			"is_shared": shared,
			"shared_at": &now,
		}).
		Scan(&palette).Error

	if err != nil {
		return PaletteDTO{}, err
	}

	return PaletteDTO{
		ID:        palette.ID,
		Name:      palette.Name,
		Palette:   palette.JsonData,
		CreatedAt: palette.CreatedAt,
		IsSystem:  palette.IsSystem,
		IsShared:  palette.IsShared,
		SharedAt:  palette.SharedAt,
	}, nil
}

func processImageWithShepardsMethod(
	img image.Image,
	paletteRGBAs []color.RGBA,
	luminosity float64,
	nearest int,
	power float64,
	maxDistanceSq float64,
) *image.RGBA {
	bounds := img.Bounds()
	height := bounds.Dy()
	out := image.NewRGBA(bounds)

	numWorkers := max(min(runtime.GOMAXPROCS(0), height), 1)
	rowsPerWorker := (height + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	for workerID := range numWorkers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			startY := bounds.Min.Y + id*rowsPerWorker
			endY := min(startY+rowsPerWorker, bounds.Max.Y)

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					originalRGBA := utils.ToRGBA(img.At(x, y))

					if originalRGBA.A == 0 {
						out.Set(x, y, color.Transparent)
						continue
					}

					if maxDistanceSq > 0 {
						if utils.NearestDistanceSquared(originalRGBA, paletteRGBAs) > maxDistanceSq {
							out.Set(x, y, originalRGBA)
							continue
						}
					}

					adjusted := utils.ApplyLuminosity(originalRGBA, luminosity)
					finalColor := utils.ShepardsMethodColor(adjusted, paletteRGBAs, nearest, power)
					out.Set(x, y, finalColor)
				}
			}
		}(workerID)
	}

	wg.Wait()
	return out
}

type ExtractResult struct {
	Palette []Color `json:"palette,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func ApplyPaletteHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file: " + err.Error()})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			_ = c.Error(err)
		}
	}()

	paletteStr := c.PostForm("palette")
	if paletteStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette is required (JSON array of hex strings or [{\"hex\":\"#RRGGBB\"}])"})
		return
	}

	var hexes []string
	if err := json.Unmarshal([]byte(paletteStr), &hexes); err != nil {
		var objs []Color
		if err2 := json.Unmarshal([]byte(paletteStr), &objs); err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid palette JSON"})
			return
		}
		for _, o := range objs {
			hexes = append(hexes, o.Hex)
		}
	}

	paletteRGBAs := make([]color.RGBA, 0, len(hexes))
	for _, h := range hexes {
		if rgba, err := utils.HexToRGBA(h); err == nil {
			paletteRGBAs = append(paletteRGBAs, rgba)
		}
	}
	if len(paletteRGBAs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Palette contained no valid colors"})
		return
	}

	luminosity := 1.0
	if s := c.PostForm("luminosity"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
			luminosity = v
		}
	}
	nearest := 30
	if s := c.PostForm("nearest"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 1 {
			nearest = v
		}
	}
	power := 4.0
	if s := c.PostForm("power"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
			power = v
		}
	}

	maxDistanceSq := 0.0
	if s := c.PostForm("maxDistance"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
			maxDistanceSq = v * v
		}
	}

	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image: " + err.Error()})
		return
	}

	out := processImageWithShepardsMethod(img, paletteRGBAs, luminosity, nearest, power, maxDistanceSq)

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode PNG: " + err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}
