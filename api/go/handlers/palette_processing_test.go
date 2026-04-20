package handlers

import (
	"image"
	"image/color"
	"testing"

	"themesmith/utils"

	"github.com/stretchr/testify/assert"
)

func TestProcessImageWithShepardsMethod(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	palette := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}}

	t.Run("Basic processing", func(t *testing.T) {
		result := processImageWithShepardsMethod(img, palette, 1.0, 2, 2.0, 0)
		assert.NotNil(t, result)
		assert.Equal(t, img.Bounds(), result.Bounds())
	})

	t.Run("With luminosity adjustment", func(t *testing.T) {
		result := processImageWithShepardsMethod(img, palette, 0.5, 2, 2.0, 0)
		assert.NotNil(t, result)
		assert.Equal(t, img.Bounds(), result.Bounds())
	})

	t.Run("With max distance threshold", func(t *testing.T) {
		result := processImageWithShepardsMethod(img, palette, 1.0, 2, 2.0, 1000.0)
		assert.NotNil(t, result)
		for y := range 4 {
			for x := range 4 {
				c := result.At(x, y)
				rgba := utils.ToRGBA(c)
				assert.NotEqual(t, color.RGBA{0, 0, 0, 0}, rgba)
			}
		}
	})

	t.Run("Transparent pixels preserved", func(t *testing.T) {
		transparentImg := image.NewRGBA(image.Rect(0, 0, 2, 2))
		transparentImg.Set(0, 0, color.RGBA{100, 100, 100, 255})
		transparentImg.Set(1, 0, color.RGBA{0, 0, 0, 0})
		transparentImg.Set(0, 1, color.RGBA{150, 150, 150, 255})
		transparentImg.Set(1, 1, color.RGBA{0, 0, 0, 0})

		result := processImageWithShepardsMethod(transparentImg, palette, 1.0, 2, 2.0, 0)

		pixel1 := result.At(1, 0)
		rgba1 := utils.ToRGBA(pixel1)
		assert.Equal(t, uint8(0), rgba1.A)

		pixel2 := result.At(1, 1)
		rgba2 := utils.ToRGBA(pixel2)
		assert.Equal(t, uint8(0), rgba2.A)
	})
}
