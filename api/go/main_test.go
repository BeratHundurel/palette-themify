package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)
// --- API Integration Tests ---

func TestSavePaletteHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/palettes", savePaletteHandler)

	t.Run("MissingName", func(t *testing.T) {
		palette := map[string]any{
			"palette": []Color{{Hex: "#FF0000"}},
		}
		paletteJSON, _ := json.Marshal(palette)

		req := httptest.NewRequest("POST", "/palettes", bytes.NewReader(paletteJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MissingPalette", func(t *testing.T) {
		palette := map[string]any{
			"name": "test",
		}
		paletteJSON, _ := json.Marshal(palette)

		req := httptest.NewRequest("POST", "/palettes", bytes.NewReader(paletteJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

type mockRoundTripper struct {
	resp *http.Response
	err  error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func TestWallhavenHandlers(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/wallhaven/search", wallhavenSearchHandler)
	router.GET("/wallhaven/w/:id", wallhavenGetWallpaperHandler)
	router.GET("/wallhaven/download", wallhavenDownloadHandler)

	makeResp := func(status int, body []byte, contentType string) *http.Response {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{contentType}},
		}
	}

	t.Run("SearchSuccess", func(t *testing.T) {
		fakeBody := []byte(`{"data": [{"id":"abc123"}]}`)
		http.DefaultTransport = &mockRoundTripper{resp: makeResp(200, fakeBody, "application/json")}

		req := httptest.NewRequest("GET", "/wallhaven/search?q=test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		var parsed map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &parsed)
		assert.NoError(t, err)
		assert.Contains(t, parsed, "data")
	})

	t.Run("GetWallpaperSuccess", func(t *testing.T) {
		fakeBody := []byte("<html>wallpaper</html>")
		http.DefaultTransport = &mockRoundTripper{resp: makeResp(200, fakeBody, "text/html")}

		req := httptest.NewRequest("GET", "/wallhaven/w/abc123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Equal(t, string(fakeBody), w.Body.String())
	})

	t.Run("DownloadSuccess", func(t *testing.T) {
		fakePNG := func() []byte {
			var b bytes.Buffer
			img := image.NewRGBA(image.Rect(0, 0, 1, 1))
			img.Set(0, 0, color.RGBA{1, 2, 3, 255})
			png.Encode(&b, img)
			return b.Bytes()
		}()

		http.DefaultTransport = &mockRoundTripper{resp: makeResp(200, fakePNG, "image/png")}

		req := httptest.NewRequest("GET", "/wallhaven/download?url=https://example.com/img.png", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Greater(t, w.Body.Len(), 0)
	})

	t.Run("DownloadInvalidURL", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/wallhaven/download", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ProxyNetworkError", func(t *testing.T) {
		http.DefaultTransport = &mockRoundTripper{err: io.ErrUnexpectedEOF}

		req := httptest.NewRequest("GET", "/wallhaven/search?q=err", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}

func TestApplyPaletteHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/apply-palette", applyPaletteHandler)

	img := createTestImage(10, 10)
	var buf bytes.Buffer
	png.Encode(&buf, img)

	palette := []Color{
		{Hex: "#FF0000"},
		{Hex: "#00FF00"},
		{Hex: "#0000FF"},
	}
	paletteJSON, _ := json.Marshal(palette)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.png")
	part.Write(buf.Bytes())
	writer.WriteField("palette", string(paletteJSON))
	writer.Close()

	req := httptest.NewRequest("POST", "/apply-palette", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Greater(t, w.Body.Len(), 0)
}

// --- Utility Functions ---

func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

func TestCreateColorHex(t *testing.T) {
	color := createColor(255, 107, 53)
	assert.Equal(t, "#FF6B35", color.Hex)

	color = createColor(0, 0, 0)
	assert.Equal(t, "#000000", color.Hex)

	color = createColor(255, 255, 255)
	assert.Equal(t, "#FFFFFF", color.Hex)
}

func TestDemoUserEmail(t *testing.T) {
	demoEmail := "demo@imagepalette.com"
	assert.Contains(t, demoEmail, "@")
	assert.Contains(t, demoEmail, "demo")
	assert.Contains(t, demoEmail, "imagepalette.com")
}

// --- Color Utility Function Tests ---

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
			result, err := hexToRGBA(tt.input)
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
			result := toRGBA(tt.input)
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
	palette := []color.RGBA{
		{255, 0, 0, 255}, // Red
		{0, 255, 0, 255}, // Green
		{0, 0, 255, 255}, // Blue
	}

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
			result := nearestDistanceSquared(tt.input, palette)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindNClosestColors(t *testing.T) {
	palette := []color.RGBA{
		{255, 0, 0, 255},   // Red
		{0, 255, 0, 255},   // Green
		{0, 0, 255, 255},   // Blue
		{255, 255, 0, 255}, // Yellow
	}

	t.Run("Find 2 closest to red-ish", func(t *testing.T) {
		input := color.RGBA{200, 50, 50, 255}
		result := findNClosestColors(input, palette, 2)
		assert.Len(t, result, 2)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, result[0].color)
	})

	t.Run("Find all colors", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := findNClosestColors(input, palette, 10)
		assert.Len(t, result, 4)
		assert.True(t, result[0].dist <= result[1].dist)
		assert.True(t, result[1].dist <= result[2].dist)
		assert.True(t, result[2].dist <= result[3].dist)
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
		assert.Equal(t, 0.0, result[0].dist)
	})
}

func TestBlendColors(t *testing.T) {
	t.Run("Equal weights", func(t *testing.T) {
		colors := []color.Color{
			color.RGBA{255, 0, 0, 255},
			color.RGBA{0, 255, 0, 255},
		}
		weights := []float64{1.0, 1.0}
		result := blendColors(colors, weights)
		assert.Equal(t, color.RGBA{128, 128, 0, 255}, result)
	})

	t.Run("Weighted blend", func(t *testing.T) {
		colors := []color.Color{
			color.RGBA{255, 0, 0, 255},
			color.RGBA{0, 0, 255, 255},
		}
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
		colors := []color.Color{
			color.RGBA{255, 0, 0, 255},
			color.RGBA{0, 255, 0, 255},
		}
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
			result := applyLuminosity(tt.input, tt.factor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShepardsMethodColor(t *testing.T) {
	palette := []color.RGBA{
		{255, 0, 0, 255}, // Red
		{0, 255, 0, 255}, // Green
		{0, 0, 255, 255}, // Blue
	}

	t.Run("Exact match returns same color", func(t *testing.T) {
		input := color.RGBA{255, 0, 0, 255}
		result := shepardsMethodColor(input, palette, 2, 2.0)
		rgba := toRGBA(result)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, rgba)
	})

	t.Run("Empty palette returns original", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := shepardsMethodColor(input, []color.RGBA{}, 2, 2.0)
		rgba := toRGBA(result)
		assert.Equal(t, input, rgba)
	})

	t.Run("Single color palette", func(t *testing.T) {
		input := color.RGBA{100, 100, 100, 255}
		singlePalette := []color.RGBA{{255, 0, 0, 255}}
		result := shepardsMethodColor(input, singlePalette, 2, 2.0)
		rgba := toRGBA(result)
		assert.Equal(t, color.RGBA{255, 0, 0, 255}, rgba)
	})

	t.Run("Blended color", func(t *testing.T) {
		input := color.RGBA{128, 128, 128, 255}
		result := shepardsMethodColor(input, palette, 3, 2.0)
		rgba := toRGBA(result)
		assert.NotEqual(t, input, rgba)
		assert.Equal(t, uint8(255), rgba.A)
	})
}

func TestColorObservationDistance(t *testing.T) {
	obs := colorObservation{0.5, 0.5, 0.5}

	t.Run("Distance to same point is 0", func(t *testing.T) {
		dist := obs.Distance([]float64{0.5, 0.5, 0.5})
		assert.Equal(t, 0.0, dist)
	})

	t.Run("Distance calculation", func(t *testing.T) {
		dist := obs.Distance([]float64{0.0, 0.0, 0.0})
		expected := 0.75
		assert.InDelta(t, expected, dist, 0.0001)
	})

	t.Run("Distance to opposite corner", func(t *testing.T) {
		dist := obs.Distance([]float64{1.0, 1.0, 1.0})
		expected := 0.75
		assert.InDelta(t, expected, dist, 0.0001)
	})
}

func TestProcessImageWithShepardsMethod(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	palette := []color.RGBA{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
	}

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
				rgba := toRGBA(c)
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
		rgba1 := toRGBA(pixel1)
		assert.Equal(t, uint8(0), rgba1.A)

		pixel2 := result.At(1, 1)
		rgba2 := toRGBA(pixel2)
		assert.Equal(t, uint8(0), rgba2.A)
	})
}

func TestExtractColors(t *testing.T) {
	sortedColors := []struct {
		dist  float64
		color color.Color
	}{
		{10.0, color.RGBA{255, 0, 0, 255}},
		{20.0, color.RGBA{0, 255, 0, 255}},
		{30.0, color.RGBA{0, 0, 255, 255}},
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
