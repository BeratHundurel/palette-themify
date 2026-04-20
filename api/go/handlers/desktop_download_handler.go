package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func DesktopDownloadHandler(c *gin.Context) {
	targetURL, ok := resolveDesktopInstallerURL(detectClientPlatform(c.GetHeader("User-Agent")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "desktop installer is not configured for this platform",
		})
		return
	}

	c.Redirect(http.StatusFound, targetURL)
}

func resolveDesktopInstallerURL(goos string) (string, bool) {
	specificKey, fallbackKey := desktopInstallerEnvKeys(goos)

	if value := strings.TrimSpace(os.Getenv(specificKey)); value != "" {
		return value, true
	}

	if value := strings.TrimSpace(os.Getenv(fallbackKey)); value != "" {
		return value, true
	}

	return "", false
}

func desktopInstallerEnvKeys(goos string) (string, string) {
	switch goos {
	case "windows":
		return "DESKTOP_INSTALLER_URL_WINDOWS", "DESKTOP_INSTALLER_URL_DEFAULT"
	case "darwin":
		return "DESKTOP_INSTALLER_URL_MACOS", "DESKTOP_INSTALLER_URL_DEFAULT"
	case "linux":
		return "DESKTOP_INSTALLER_URL_LINUX", "DESKTOP_INSTALLER_URL_DEFAULT"
	default:
		return fmt.Sprintf("DESKTOP_INSTALLER_URL_%s", strings.ToUpper(goos)), "DESKTOP_INSTALLER_URL_DEFAULT"
	}
}

func detectClientPlatform(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "windows"):
		return "windows"
	case strings.Contains(ua, "macintosh"), strings.Contains(ua, "mac os"), strings.Contains(ua, "darwin"):
		return "darwin"
	case strings.Contains(ua, "linux"), strings.Contains(ua, "x11"):
		return "linux"
	default:
		return "unknown"
	}
}
