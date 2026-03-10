package utils

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateColorHex(t *testing.T) {
	color := createColor(255, 107, 53)
	assert.Equal(t, "#FF6B35", color.Hex)

	color = createColor(0, 0, 0)
	assert.Equal(t, "#000000", color.Hex)

	color = createColor(255, 255, 255)
	assert.Equal(t, "#FFFFFF", color.Hex)
}

func TestHexToRGBA(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    color.RGBA
		expectError bool
	}{
		{"Valid with hash", "#FF6B35", color.RGBA{255, 107, 53, 255}, false},
		{"Valid without hash", "FF6B35", color.RGBA{255, 107, 53, 255}, false},
		{"Black", "#000000", color.RGBA{0, 0, 0, 255}, false},
		{"White", "#FFFFFF", color.RGBA{255, 255, 255, 255}, false},
		{"With spaces", "  #AABBCC  ", color.RGBA{170, 187, 204, 255}, false},
		{"Empty string", "", color.RGBA{}, true},
		{"Invalid length", "#FFF", color.RGBA{}, true},
		{"Invalid hex", "#GGGGGG", color.RGBA{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HexToRGBA(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestToRGBA(t *testing.T) {
	tests := []struct {
		name     string
		input    color.Color
		expected color.RGBA
	}{
		{"RGBA color", color.RGBA{100, 150, 200, 255}, color.RGBA{100, 150, 200, 255}},
		{"NRGBA color", color.NRGBA{50, 100, 150, 200}, color.RGBA{39, 78, 118, 200}},
		{"Black", color.RGBA{0, 0, 0, 255}, color.RGBA{0, 0, 0, 255}},
		{"White", color.RGBA{255, 255, 255, 255}, color.RGBA{255, 255, 255, 255}},
		{"Transparent", color.RGBA{0, 0, 0, 0}, color.RGBA{0, 0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToRGBA(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorDistanceSquared(t *testing.T) {
	tests := []struct {
		name     string
		c1       color.RGBA
		c2       color.RGBA
		expected float64
	}{
		{"Identical colors", color.RGBA{100, 100, 100, 255}, color.RGBA{100, 100, 100, 255}, 0},
		{"Black and white", color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}, 195075},
		{"Red to green", color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}, 130050},
		{"Small difference", color.RGBA{100, 100, 100, 255}, color.RGBA{101, 101, 101, 255}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorDistanceSquared(tt.c1, tt.c2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNearestDistanceSquared(t *testing.T) {
	palette := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}}

	tests := []struct {
		name     string
		input    color.RGBA
		expected float64
	}{
		{"Exact match red", color.RGBA{255, 0, 0, 255}, 0},
		{"Close to red", color.RGBA{250, 5, 5, 255}, 75},
		{"Close to green", color.RGBA{5, 250, 5, 255}, 75},
		{"Gray - equidistant", color.RGBA{128, 128, 128, 255}, 48897},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NearestDistanceSquared(tt.input, palette)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindNClosestColors(t *testing.T) {
	palette := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}, {255, 255, 0, 255}}

	t.Run("Find 2 closest to red-ish", func(t *testing.T) {
		input := color.RGBA{200, 50, 50, 255}
		result := findNClosestColors(input, palette, 2)
		assert.Len(t, result, 2)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, result[0].Color)
	})

	t.Run("Find all colors", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := findNClosestColors(input, palette, 10)
		assert.Len(t, result, 4)
		assert.True(t, result[0].Distance <= result[1].Distance)
		assert.True(t, result[1].Distance <= result[2].Distance)
		assert.True(t, result[2].Distance <= result[3].Distance)
	})

	t.Run("Empty palette", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := findNClosestColors(input, []color.RGBA{}, 2)
		assert.Nil(t, result)
	})

	t.Run("Exact match", func(t *testing.T) {
		input := color.RGBA{255, 0, 0, 255}
		result := findNClosestColors(input, palette, 2)
		assert.Len(t, result, 2)
		assert.Equal(t, 0.0, result[0].Distance)
	})
}

func TestBlendColors(t *testing.T) {
	t.Run("Equal weights", func(t *testing.T) {
		colors := []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}}
		weights := []float64{1.0, 1.0}
		result := blendColors(colors, weights)
		assert.Equal(t, color.RGBA{128, 128, 0, 255}, result)
	})

	t.Run("Weighted blend", func(t *testing.T) {
		colors := []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}}
		weights := []float64{3.0, 1.0}
		result := blendColors(colors, weights)
		assert.Equal(t, color.RGBA{191, 0, 64, 255}, result)
	})

	t.Run("Single color", func(t *testing.T) {
		colors := []color.Color{color.RGBA{100, 150, 200, 255}}
		weights := []float64{1.0}
		result := blendColors(colors, weights)
		assert.Equal(t, color.RGBA{100, 150, 200, 255}, result)
	})

	t.Run("Empty colors", func(t *testing.T) {
		result := blendColors([]color.Color{}, []float64{})
		assert.Equal(t, color.RGBA{}, result)
	})

	t.Run("Mismatched lengths", func(t *testing.T) {
		colors := []color.Color{color.RGBA{255, 0, 0, 255}}
		weights := []float64{1.0, 2.0}
		result := blendColors(colors, weights)
		assert.Equal(t, color.RGBA{}, result)
	})

	t.Run("Zero total weight", func(t *testing.T) {
		colors := []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}}
		weights := []float64{0.0, 0.0}
		result := blendColors(colors, weights)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, result)
	})
}

func TestApplyLuminosity(t *testing.T) {
	tests := []struct {
		name     string
		input    color.RGBA
		factor   float64
		expected color.RGBA
	}{
		{"No change", color.RGBA{100, 150, 200, 255}, 1.0, color.RGBA{100, 150, 200, 255}},
		{"Darken by half", color.RGBA{100, 150, 200, 255}, 0.5, color.RGBA{50, 75, 100, 255}},
		{"Brighten by 1.5x", color.RGBA{100, 100, 100, 255}, 1.5, color.RGBA{150, 150, 150, 255}},
		{"Clamp to 255", color.RGBA{200, 200, 200, 255}, 2.0, color.RGBA{255, 255, 255, 255}},
		{"Clamp to 0", color.RGBA{50, 50, 50, 255}, 0.0, color.RGBA{0, 0, 0, 255}},
		{"Alpha preserved", color.RGBA{100, 100, 100, 128}, 0.5, color.RGBA{50, 50, 50, 128}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyLuminosity(tt.input, tt.factor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShepardsMethodColor(t *testing.T) {
	palette := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}}

	t.Run("Exact match returns same color", func(t *testing.T) {
		input := color.RGBA{255, 0, 0, 255}
		result := ShepardsMethodColor(input, palette, 2, 2.0)
		rgba := ToRGBA(result)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, rgba)
	})

	t.Run("Empty palette returns original", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := ShepardsMethodColor(input, []color.RGBA{}, 2, 2.0)
		rgba := ToRGBA(result)
		assert.Equal(t, input, rgba)
	})

	t.Run("Single color palette", func(t *testing.T) {
		input := color.RGBA{100, 100, 100, 255}
		singlePalette := []color.RGBA{{255, 0, 0, 255}}
		result := ShepardsMethodColor(input, singlePalette, 2, 2.0)
		rgba := ToRGBA(result)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, rgba)
	})

	t.Run("Blended color", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := ShepardsMethodColor(input, palette, 3, 2.0)
		rgba := ToRGBA(result)
		assert.NotEqual(t, input, rgba)
		assert.Equal(t, uint8(255), rgba.A)
	})
}

func TestExtractColors(t *testing.T) {
	sortedColors := []weightedColor{
		{Distance: 10.0, Color: color.RGBA{255, 0, 0, 255}},
		{Distance: 20.0, Color: color.RGBA{0, 255, 0, 255}},
		{Distance: 30.0, Color: color.RGBA{0, 0, 255, 255}},
	}

	result := extractColors(sortedColors)
	assert.Len(t, result, 3)
	assert.Equal(t, color.RGBA{255, 0, 0, 255}, result[0])
	assert.Equal(t, color.RGBA{0, 255, 0, 255}, result[1])
	assert.Equal(t, color.RGBA{0, 0, 255, 255}, result[2])
}

func TestMinInt(t *testing.T) {
	assert.Equal(t, 1, minInt(1, 2))
	assert.Equal(t, 1, minInt(2, 1))
	assert.Equal(t, 0, minInt(0, 5))
	assert.Equal(t, -5, minInt(-5, 10))
	assert.Equal(t, 5, minInt(5, 5))
}
