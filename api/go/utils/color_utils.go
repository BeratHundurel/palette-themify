package utils

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"

	"image-to-palette/model"
)

type weightedColor struct {
	Distance float64
	Color    color.Color
}

func HexToRGBA(s string) (color.RGBA, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.RGBA{}, fmt.Errorf("empty color")
	}
	if s[0] == '#' {
		s = s[1:]
	}
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color")
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{
		R: uint8(v >> 16),
		G: uint8((v >> 8) & 0xFF),
		B: uint8(v & 0xFF),
		A: 255,
	}, nil
}

func ToRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func colorDistanceSquared(c1, c2 color.RGBA) float64 {
	dr := float64(c1.R) - float64(c2.R)
	dg := float64(c1.G) - float64(c2.G)
	db := float64(c1.B) - float64(c2.B)
	return dr*dr + dg*dg + db*db
}

func NearestDistanceSquared(c color.RGBA, paletteRGBAs []color.RGBA) float64 {
	min := math.MaxFloat64
	for _, p := range paletteRGBAs {
		d := colorDistanceSquared(c, p)
		if d < min {
			min = d
		}
	}
	return min
}

func findNClosestColors(originalRGBA color.RGBA, paletteRGBAs []color.RGBA, n int) []weightedColor {
	if len(paletteRGBAs) == 0 {
		return nil
	}
	distances := make([]weightedColor, 0, len(paletteRGBAs))
	for _, pRGBA := range paletteRGBAs {
		distances = append(distances, weightedColor{
			Distance: colorDistanceSquared(originalRGBA, pRGBA),
			Color:    pRGBA,
		})
	}
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].Distance < distances[j].Distance
	})
	if n > len(distances) {
		n = len(distances)
	}
	return distances[:n]
}

func extractColors(sortedColors []weightedColor) []color.Color {
	colors := make([]color.Color, len(sortedColors))
	for i, item := range sortedColors {
		colors[i] = item.Color
	}
	return colors
}

func blendColors(colors []color.Color, weights []float64) color.RGBA {
	if len(colors) == 0 || len(colors) != len(weights) {
		return color.RGBA{}
	}
	var sumR, sumG, sumB float64
	var totalWeight float64
	for i := range colors {
		rgba := ToRGBA(colors[i])
		sumR += float64(rgba.R) * weights[i]
		sumG += float64(rgba.G) * weights[i]
		sumB += float64(rgba.B) * weights[i]
		totalWeight += weights[i]
	}
	if totalWeight == 0 {
		return ToRGBA(colors[0])
	}
	return color.RGBA{
		R: uint8(math.Round(sumR / totalWeight)),
		G: uint8(math.Round(sumG / totalWeight)),
		B: uint8(math.Round(sumB / totalWeight)),
		A: 255,
	}
}

func ApplyLuminosity(c color.RGBA, factor float64) color.RGBA {
	r := uint8(math.Max(0, math.Min(255, float64(c.R)*factor)))
	g := uint8(math.Max(0, math.Min(255, float64(c.G)*factor)))
	b := uint8(math.Max(0, math.Min(255, float64(c.B)*factor)))
	return color.RGBA{R: r, G: g, B: b, A: c.A}
}

func ShepardsMethodColor(originalRGBA color.RGBA, paletteRGBAs []color.RGBA, nearest int, power float64) color.Color {
	closest := findNClosestColors(originalRGBA, paletteRGBAs, nearest)
	if len(closest) == 0 {
		return originalRGBA
	}
	if len(closest) == 1 || closest[0].Distance == 0 {
		return closest[0].Color
	}

	weights := make([]float64, len(closest))
	var totalWeight float64
	for i, c := range closest {
		if c.Distance == 0 {
			return c.Color
		}
		weight := 1.0 / math.Pow(math.Sqrt(c.Distance), power)
		weights[i] = weight
		totalWeight += weight
	}
	if totalWeight == 0 {
		return closest[0].Color
	}
	return blendColors(extractColors(closest), weights)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func createColor(r, g, b uint8) model.Color {
	return model.Color{
		Hex: fmt.Sprintf("#%02X%02X%02X", r, g, b),
	}
}
