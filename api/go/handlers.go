package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Color struct {
	Hex string `json:"hex"`
}

type ExtractResult struct {
	Palette []Color `json:"palette,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func applyPaletteHandler(c *gin.Context) {
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
	defer file.Close()

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
		if rgba, err := hexToRGBA(h); err == nil {
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
