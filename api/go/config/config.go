package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	defaultHTTPHost = "0.0.0.0"
	defaultHTTPPort = 8088
)

type Config struct {
	Environment    string
	HTTPAddress    string
	AllowedOrigins []string
	ServiceName    string
	ServiceVersion string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	environment := env("APP_ENV", "development")
	port, err := strconv.Atoi(env("PORT", strconv.Itoa(defaultHTTPPort)))
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
	}

	address := strings.TrimSpace(os.Getenv("HTTP_ADDRESS"))
	if address == "" {
		address = net.JoinHostPort(env("HTTP_HOST", defaultHTTPHost), strconv.Itoa(port))
	}

	origins, err := parseOrigins(env("ALLOWED_ORIGINS", defaultOrigins(environment)))
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Environment:    environment,
		HTTPAddress:    address,
		AllowedOrigins: origins,
		ServiceName:    env("OTEL_SERVICE_NAME", "themesmith-api"),
		ServiceVersion: env("APP_VERSION", "dev"),
	}

	if err := validateProduction(config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func defaultOrigins(environment string) string {
	if environment == "production" {
		return ""
	}
	return "http://localhost:5173,http://127.0.0.1:5173,http://wails.localhost:9245"
}

func parseOrigins(raw string) ([]string, error) {
	var origins []string
	for value := range strings.SplitSeq(raw, ",") {
		origin := strings.TrimSpace(value)
		if origin == "" {
			continue
		}

		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, fmt.Errorf("ALLOWED_ORIGINS contains an invalid origin %q", origin)
		}
		if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
			return nil, fmt.Errorf("ALLOWED_ORIGINS entries must contain only scheme and host: %q", origin)
		}

		origins = append(origins, fmt.Sprintf("%s://%s", strings.ToLower(parsed.Scheme), strings.ToLower(parsed.Host)))
	}
	return origins, nil
}

func validateProduction(config Config) error {
	if config.Environment != "production" {
		return nil
	}

	if len(config.AllowedOrigins) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS is required when APP_ENV=production")
	}

	frontendOrigins, err := parseOrigins(strings.TrimSpace(os.Getenv("FRONTEND_URL")))
	if err != nil || len(frontendOrigins) != 1 {
		return fmt.Errorf("FRONTEND_URL must be one absolute HTTP or HTTPS origin when APP_ENV=production")
	}
	frontendAllowed := false
	for _, allowedOrigin := range config.AllowedOrigins {
		if frontendOrigins[0] == allowedOrigin {
			frontendAllowed = true
			break
		}
	}
	if !frontendAllowed {
		return fmt.Errorf("FRONTEND_URL must also be present in ALLOWED_ORIGINS when APP_ENV=production")
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) < 32 || strings.Contains(strings.ToLower(jwtSecret), "change-me") {
		return fmt.Errorf("JWT_SECRET must be a non-default value of at least 32 characters when APP_ENV=production")
	}

	databasePassword := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	if databasePassword == "" || databasePassword == "password" || strings.Contains(strings.ToLower(databasePassword), "change-me") {
		return fmt.Errorf("DB_PASSWORD must be set to a non-default value when APP_ENV=production")
	}

	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	if (googleClientID == "") != (googleClientSecret == "") {
		return fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set together")
	}
	if googleClientID != "" {
		redirectURL := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL"))
		parsedRedirect, err := url.Parse(redirectURL)
		if err != nil || parsedRedirect.Scheme != "https" || parsedRedirect.Host == "" {
			return fmt.Errorf("GOOGLE_REDIRECT_URL must be an absolute HTTPS URL when Google authentication is enabled in production")
		}
	}

	return nil
}
