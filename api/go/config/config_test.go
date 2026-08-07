package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "8090")
	t.Setenv("HTTP_ADDRESS", "")
	t.Setenv("HTTP_HOST", "127.0.0.1")
	t.Setenv("ALLOWED_ORIGINS", "")

	config, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8090", config.HTTPAddress)
	assert.Contains(t, config.AllowedOrigins, "http://localhost:5173")
}

func TestLoadRejectsProductionDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://themesmith.example")
	t.Setenv("FRONTEND_URL", "https://themesmith.example")
	t.Setenv("JWT_SECRET", "change-me")
	t.Setenv("DB_PASSWORD", "change-me")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoadAcceptsProductionConfiguration(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://themesmith.example")
	t.Setenv("FRONTEND_URL", "https://themesmith.example")
	t.Setenv("JWT_SECRET", "a-production-secret-that-is-long-enough")
	t.Setenv("DB_PASSWORD", "a-database-password")

	config, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://themesmith.example"}, config.AllowedOrigins)
}

func TestLoadRejectsOriginPaths(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("ALLOWED_ORIGINS", "https://themesmith.example/path")

	_, err := Load()
	require.Error(t, err)
}
