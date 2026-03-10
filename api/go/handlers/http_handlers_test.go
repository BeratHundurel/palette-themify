package handlers

import (
	"bytes"
	"encoding/json"
	"image"
	"image-to-palette/model"
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

func TestSavePaletteHandler_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/palettes", SavePaletteHandler)

	t.Run("MissingName", func(t *testing.T) {
		palette := map[string]any{
			"palette": []model.Color{{Hex: "#FF0000"}},
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

func TestWallhavenHandlers(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/wallhaven/search", WallhavenSearchHandler)
	router.GET("/wallhaven/w/:id", WallhavenGetWallpaperHandler)
	router.GET("/wallhaven/download", WallhavenDownloadHandler)

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
			err := png.Encode(&b, img)
			assert.NoError(t, err)
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
	router.POST("/apply-palette", ApplyPaletteHandler)

	malformedBody := &bytes.Buffer{}
	malformedWriter := multipart.NewWriter(malformedBody)
	if err := malformedWriter.WriteField("palette", "invalid-json"); err != nil {
		t.Fatalf("write malformed palette: %v", err)
	}
	if err := malformedWriter.Close(); err != nil {
		t.Fatalf("close malformed writer: %v", err)
	}
	malformedReq := httptest.NewRequest("POST", "/apply-palette", malformedBody)
	malformedReq.Header.Set("Content-Type", malformedWriter.FormDataContentType())
	malformedResp := httptest.NewRecorder()
	router.ServeHTTP(malformedResp, malformedReq)
	assert.Equal(t, http.StatusBadRequest, malformedResp.Code)

	img := createTestImage(10, 10)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	palette := []model.Color{{Hex: "#FF0000"}, {Hex: "#00FF00"}, {Hex: "#0000FF"}}
	paletteJSON, _ := json.Marshal(palette)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.png")
	if _, err := part.Write(buf.Bytes()); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.WriteField("palette", string(paletteJSON)); err != nil {
		t.Fatalf("write palette field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/apply-palette", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Greater(t, w.Body.Len(), 0)
}
