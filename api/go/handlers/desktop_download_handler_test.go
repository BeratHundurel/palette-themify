package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDetectClientPlatform(t *testing.T) {
	assert.Equal(t, "windows", detectClientPlatform("Mozilla/5.0 (Windows NT 10.0; Win64; x64)"))
	assert.Equal(t, "darwin", detectClientPlatform("Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4)"))
	assert.Equal(t, "linux", detectClientPlatform("Mozilla/5.0 (X11; Linux x86_64)"))
	assert.Equal(t, "unknown", detectClientPlatform("Mozilla/5.0 (PlayStation 5)"))
}

func TestResolveDesktopInstallerURL(t *testing.T) {
	t.Setenv("DESKTOP_INSTALLER_URL_WINDOWS", "https://downloads.example.com/themesmith/windows/setup.exe")
	t.Setenv("DESKTOP_INSTALLER_URL_DEFAULT", "https://downloads.example.com/themesmith/latest")

	url, ok := resolveDesktopInstallerURL("windows")
	assert.True(t, ok)
	assert.Equal(t, "https://downloads.example.com/themesmith/windows/setup.exe", url)

	url, ok = resolveDesktopInstallerURL("linux")
	assert.True(t, ok)
	assert.Equal(t, "https://downloads.example.com/themesmith/latest", url)
}

func TestDesktopDownloadHandler_Redirects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/desktop/download", DesktopDownloadHandler)

	t.Setenv("DESKTOP_INSTALLER_URL_WINDOWS", "https://downloads.example.com/themesmith/windows/setup.exe")

	req := httptest.NewRequest(http.MethodGet, "/desktop/download", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://downloads.example.com/themesmith/windows/setup.exe", w.Header().Get("Location"))
}

func TestDesktopDownloadHandler_NotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/desktop/download", DesktopDownloadHandler)

	_ = os.Unsetenv("DESKTOP_INSTALLER_URL_WINDOWS")
	_ = os.Unsetenv("DESKTOP_INSTALLER_URL_MACOS")
	_ = os.Unsetenv("DESKTOP_INSTALLER_URL_LINUX")
	_ = os.Unsetenv("DESKTOP_INSTALLER_URL_DEFAULT")

	req := httptest.NewRequest(http.MethodGet, "/desktop/download", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "desktop installer is not configured")
}
