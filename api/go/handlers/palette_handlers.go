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
)

type PaletteData struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Palette   []model.Color `json:"palette"`
	CreatedAt time.Time     `json:"createdAt"`
	IsSystem  bool          `json:"isSystem"`
	IsShared  bool          `json:"isShared"`
	SharedAt  *time.Time    `json:"sharedAt"`
}

type SavePaletteRequest struct {
	Name    string        `json:"name" binding:"required"`
	Palette []model.Color `json:"palette" binding:"required"`
}

type DeletePalettesRequest struct {
	IDs []string `json:"ids"`
}

type GetPalettesResponse struct {
	Palettes []PaletteData `json:"palettes"`
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Palette saved successfully",
		"name":    req.Name,
	})
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

	c.JSON(http.StatusOK, gin.H{"message": "Palette deleted successfully"})
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

	deletedCount, err := deleteUserPalettes(userID, req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Palettes deleted successfully", "deleted": deletedCount})
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

	c.JSON(http.StatusOK, gin.H{"message": "Palette shared successfully", "palette": palette})
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

	c.JSON(http.StatusOK, gin.H{"message": "Palette unshared successfully", "palette": palette})
}

func saveUserPalette(userID uint, name string, palette []model.Color) error {
	if db.DB == nil {
		return fmt.Errorf("database not available")
	}

	paletteJSON, err := json.Marshal(palette)
	if err != nil {
		return err
	}

	dbPalette := model.Palette{
		UserID:   &userID,
		JsonData: string(paletteJSON),
		Name:     name,
	}

	return db.DB.Create(&dbPalette).Error
}

func getUserPalettes(userID uint) ([]PaletteData, error) {
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

	palettes := make([]PaletteData, len(dbPalettes))
	for i, dbPalette := range dbPalettes {
		var colors []model.Color
		if err := json.Unmarshal([]byte(dbPalette.JsonData), &colors); err != nil {
			continue
		}

		palettes[i] = PaletteData{
			ID:        fmt.Sprintf("%d", dbPalette.ID),
			Name:      dbPalette.Name,
			Palette:   colors,
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

	var palette model.Palette
	if err := db.DB.Where("id = ? AND user_id = ?", paletteID, userID).First(&palette).Error; err != nil {
		return fmt.Errorf("palette not found or unauthorized")
	}

	if palette.IsSystem {
		return fmt.Errorf("cannot delete system palettes")
	}

	result := db.DB.Delete(&palette)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func deleteUserPalettes(userID uint, paletteIDs []string) (int64, error) {
	if db.DB == nil {
		return 0, fmt.Errorf("database not available")
	}

	result := db.DB.Where("id IN ? AND user_id = ? AND is_system = ?", paletteIDs, userID, false).Delete(&model.Palette{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func setPaletteShared(userID uint, paletteID string, shared bool) (PaletteData, error) {
	if db.DB == nil {
		return PaletteData{}, fmt.Errorf("database not available")
	}

	var dbPalette model.Palette
	if err := db.DB.Where("id = ? AND user_id = ?", paletteID, userID).First(&dbPalette).Error; err != nil {
		return PaletteData{}, fmt.Errorf("palette not found or unauthorized")
	}

	dbPalette.IsShared = shared
	if shared {
		now := time.Now().UTC()
		dbPalette.SharedAt = &now
	} else {
		dbPalette.SharedAt = nil
	}
	dbPalette.UpdatedAt = time.Now().UTC()

	if err := db.DB.Save(&dbPalette).Error; err != nil {
		return PaletteData{}, err
	}

	var colors []model.Color
	if err := json.Unmarshal([]byte(dbPalette.JsonData), &colors); err != nil {
		return PaletteData{}, err
	}

	return PaletteData{
		ID:        fmt.Sprintf("%d", dbPalette.ID),
		Name:      dbPalette.Name,
		Palette:   colors,
		CreatedAt: dbPalette.CreatedAt,
		IsSystem:  dbPalette.IsSystem,
		IsShared:  dbPalette.IsShared,
		SharedAt:  dbPalette.SharedAt,
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
	Palette []model.Color `json:"palette,omitempty"`
	Error   string        `json:"error,omitempty"`
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
		var objs []model.Color
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
