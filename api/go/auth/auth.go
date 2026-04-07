package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"image-to-palette/db"
	"image-to-palette/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token   string     `json:"token"`
	User    model.User `json:"user"`
	Message string     `json:"message"`
}

var googleOAuthConfig *oauth2.Config

type googleAuthMode string

const (
	googleAuthModeWeb     googleAuthMode = "web"
	googleAuthModeDesktop googleAuthMode = "desktop"
	googleAuthSessionTTL                 = 10 * time.Minute
)

type googleAuthSession struct {
	Mode        googleAuthMode
	RedirectURL string
	ExpiresAt   time.Time
	Token       string
	User        *model.User
	Error       string
}

var (
	googleAuthSessions   = map[string]*googleAuthSession{}
	googleAuthSessionsMu sync.Mutex
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getGoogleOAuthConfig() *oauth2.Config {
	if googleOAuthConfig != nil {
		return googleOAuthConfig
	}

	googleOAuthConfig = &oauth2.Config{
		ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
		RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8088/auth/google/callback"),
	}

	return googleOAuthConfig
}

func getAllowedAuthOrigins() map[string]struct{} {
	origins := map[string]struct{}{}

	for _, rawOrigin := range []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://wails.localhost:9245",
		getEnv("FRONTEND_URL", "http://localhost:5173"),
	} {
		if origin, err := normalizeAndValidateOrigin(rawOrigin); err == nil {
			origins[origin] = struct{}{}
		}
	}

	for rawOrigin := range strings.SplitSeq(getEnv("ALLOWED_AUTH_ORIGINS", ""), ",") {
		if origin, err := normalizeAndValidateOrigin(rawOrigin); err == nil {
			origins[origin] = struct{}{}
		}
	}

	return origins
}

func normalizeAndValidateOrigin(rawOrigin string) (string, error) {
	trimmed := strings.TrimSpace(rawOrigin)
	if trimmed == "" {
		return "", errors.New("origin is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid origin")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid origin scheme")
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("origin host is required")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("origin must not include a path")
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return "", fmt.Errorf("origin must not include credentials, query, or fragment")
	}

	return fmt.Sprintf("%s://%s", strings.ToLower(parsed.Scheme), strings.ToLower(parsed.Host)), nil
}

func resolveGoogleRedirectURL(rawOrigin string) (string, error) {
	origin := rawOrigin
	if strings.TrimSpace(origin) == "" {
		origin = getEnv("FRONTEND_URL", "http://localhost:5173")
	}

	normalizedOrigin, err := normalizeAndValidateOrigin(origin)
	if err != nil {
		return "", fmt.Errorf("invalid origin")
	}

	if _, ok := getAllowedAuthOrigins()[normalizedOrigin]; !ok {
		return "", fmt.Errorf("origin is not allowed")
	}

	return normalizedOrigin + "/auth/google/callback", nil
}

func createGoogleAuthSession(mode googleAuthMode, redirectURL string) (string, time.Time, error) {
	state, err := generateSecureToken(32)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(googleAuthSessionTTL)

	googleAuthSessionsMu.Lock()
	defer googleAuthSessionsMu.Unlock()

	cleanupExpiredGoogleSessionsLocked(time.Now())
	googleAuthSessions[state] = &googleAuthSession{
		Mode:        mode,
		RedirectURL: redirectURL,
		ExpiresAt:   expiresAt,
	}

	return state, expiresAt, nil
}

func cleanupExpiredGoogleSessionsLocked(now time.Time) {
	for key, session := range googleAuthSessions {
		if now.After(session.ExpiresAt) {
			delete(googleAuthSessions, key)
		}
	}
}

func generateSecureToken(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWTToken(user model.User) (string, error) {
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "image-to-palette",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWTToken(tokenString string) (*Claims, error) {
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := ValidateJWTToken(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user not found in context")
	}

	uid, ok := userID.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type")
	}

	return uid, nil
}

func GetUserFromRequest(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header required")
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		return 0, errors.New("invalid authorization header format")
	}

	claims, err := ValidateJWTToken(bearerToken[1])
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if len(username) < 3 || len(username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 3 and 50 characters"})
		return
	}

	email := buildLocalEmailForUsername(username)

	var existingUser model.User
	if err := db.DB.Where("name = ?", username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
		return
	}

	if err := db.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := model.User{
		Name:         username,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := GenerateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusCreated, AuthResponse{
		Token:   token,
		User:    user,
		Message: "User registered successfully",
	})
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if len(username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be 50 characters or less"})
		return
	}

	var user model.User
	if err := db.DB.Where("name = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := GenerateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, AuthResponse{
		Token:   token,
		User:    user,
		Message: "Login successful",
	})
}

func buildLocalEmailForUsername(username string) string {
	normalized := strings.ToLower(strings.TrimSpace(username))
	return normalized + "@local.image-to-palette"
}

func GetMeHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func ChangePasswordHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	if err := db.DB.Model(&user).Update("password_hash", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func GoogleLoginHandler(c *gin.Context) {
	config := getGoogleOAuthConfig()

	if config.ClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth not configured"})
		return
	}

	mode := strings.ToLower(strings.TrimSpace(c.DefaultQuery("mode", string(googleAuthModeWeb))))

	var (
		authMode    googleAuthMode
		redirectURL string
	)

	switch mode {
	case string(googleAuthModeWeb):
		authMode = googleAuthModeWeb

		resolvedRedirectURL, err := resolveGoogleRedirectURL(c.Query("origin"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		redirectURL = resolvedRedirectURL
	case string(googleAuthModeDesktop):
		authMode = googleAuthModeDesktop
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth mode"})
		return
	}

	state, expiresAt, err := createGoogleAuthSession(authMode, redirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Google OAuth session"})
		return
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	payload := gin.H{
		"url":        authURL,
		"expires_at": expiresAt.UTC().Format(time.RFC3339),
	}

	if authMode == googleAuthModeDesktop {
		payload["session_id"] = state
	}

	c.JSON(http.StatusOK, payload)
}

func GoogleCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code parameter"})
		return
	}

	state := strings.TrimSpace(c.Query("state"))
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state parameter"})
		return
	}

	googleAuthSessionsMu.Lock()
	cleanupExpiredGoogleSessionsLocked(time.Now())
	session, exists := googleAuthSessions[state]
	if !exists {
		googleAuthSessionsMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OAuth session"})
		return
	}
	googleAuthSessionsMu.Unlock()

	config := getGoogleOAuthConfig()

	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		recordGoogleSessionError(state, session.Mode, "Failed to exchange code for token")
		if session.Mode == googleAuthModeDesktop {
			renderGoogleAuthResultPage(c, "Sign in failed", "Could not verify Google sign in. Return to the app and try again.")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	client := config.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		recordGoogleSessionError(state, session.Mode, "Failed to get user info")
		if session.Mode == googleAuthModeDesktop {
			renderGoogleAuthResultPage(c, "Sign in failed", "Could not fetch your Google profile. Return to the app and try again.")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		recordGoogleSessionError(state, session.Mode, "Failed to parse user info")
		if session.Mode == googleAuthModeDesktop {
			renderGoogleAuthResultPage(c, "Sign in failed", "Google returned an invalid profile response. Return to the app and try again.")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	if !userInfo.VerifiedEmail {
		recordGoogleSessionError(state, session.Mode, "Google email not verified")
		if session.Mode == googleAuthModeDesktop {
			renderGoogleAuthResultPage(c, "Sign in failed", "Your Google email is not verified. Verify it and try again.")
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "Google email not verified"})
		return
	}

	user, err := findOrCreateGoogleUser(userInfo)
	if err != nil {
		recordGoogleSessionError(state, session.Mode, "Failed to create user")
		if session.Mode == googleAuthModeDesktop {
			renderGoogleAuthResultPage(c, "Sign in failed", "Could not finalize your account. Return to the app and try again.")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	jwtToken, err := GenerateJWTToken(*user)
	if err != nil {
		recordGoogleSessionError(state, session.Mode, "Failed to generate token")
		if session.Mode == googleAuthModeDesktop {
			renderGoogleAuthResultPage(c, "Sign in failed", "Could not create your session token. Return to the app and try again.")
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	safeUser := *user
	safeUser.PasswordHash = ""

	if session.Mode == googleAuthModeDesktop {
		googleAuthSessionsMu.Lock()
		if currentSession, ok := googleAuthSessions[state]; ok {
			currentSession.Token = jwtToken
			currentSession.User = &safeUser
			currentSession.Error = ""
		}
		googleAuthSessionsMu.Unlock()

		renderGoogleAuthResultPage(c, "Sign in complete", "Google sign in succeeded. You can close this tab and return to the desktop app.")
		return
	}

	googleAuthSessionsMu.Lock()
	delete(googleAuthSessions, state)
	googleAuthSessionsMu.Unlock()

	fragment := url.Values{}
	fragment.Set("token", jwtToken)
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s#%s", session.RedirectURL, fragment.Encode()))
}

func GoogleDesktopStatusHandler(c *gin.Context) {
	sessionID := strings.TrimSpace(c.Query("session_id"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id parameter"})
		return
	}

	googleAuthSessionsMu.Lock()
	defer googleAuthSessionsMu.Unlock()

	cleanupExpiredGoogleSessionsLocked(time.Now())

	session, exists := googleAuthSessions[sessionID]
	if !exists || session.Mode != googleAuthModeDesktop {
		c.JSON(http.StatusNotFound, gin.H{"error": "OAuth session not found"})
		return
	}

	if session.Error != "" {
		errMessage := session.Error
		delete(googleAuthSessions, sessionID)
		c.JSON(http.StatusOK, gin.H{"status": "error", "error": errMessage})
		return
	}

	if session.Token == "" || session.User == nil {
		c.JSON(http.StatusOK, gin.H{"status": "pending"})
		return
	}

	authResponse := AuthResponse{
		Token:   session.Token,
		User:    *session.User,
		Message: "Login successful",
	}

	delete(googleAuthSessions, sessionID)
	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
		"auth":   authResponse,
	})
}

func recordGoogleSessionError(sessionID string, mode googleAuthMode, message string) {
	if mode != googleAuthModeDesktop {
		googleAuthSessionsMu.Lock()
		delete(googleAuthSessions, sessionID)
		googleAuthSessionsMu.Unlock()
		return
	}

	googleAuthSessionsMu.Lock()
	if currentSession, ok := googleAuthSessions[sessionID]; ok {
		currentSession.Error = message
	}
	googleAuthSessionsMu.Unlock()
}

func renderGoogleAuthResultPage(c *gin.Context, title string, message string) {
	html := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s</title>
  <style>
    :root {
      color-scheme: dark;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    body {
      margin: 0;
      min-height: 100svh;
      display: grid;
      place-items: center;
      background: radial-gradient(circle at top, #1f2937, #0b1020 60%%);
      color: #e5e7eb;
    }
    main {
      width: min(92vw, 480px);
      border: 1px solid rgba(148, 163, 184, 0.35);
      border-radius: 14px;
      padding: 24px;
      background: rgba(15, 23, 42, 0.9);
      box-shadow: 0 18px 48px rgba(2, 6, 23, 0.5);
    }
    h1 {
      margin: 0 0 10px;
      font-size: 1.25rem;
    }
    p {
      margin: 0;
      line-height: 1.6;
      color: #cbd5e1;
    }
  </style>
</head>
<body>
  <main>
    <h1>%s</h1>
    <p>%s</p>
  </main>
</body>
</html>`, title, title, message)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func findOrCreateGoogleUser(userInfo GoogleUserInfo) (*model.User, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	var user model.User
	if err := db.DB.Where("google_id = ?", userInfo.ID).First(&user).Error; err == nil {
		if userInfo.Picture != "" && user.AvatarURL != userInfo.Picture {
			db.DB.Model(&user).Update("avatar_url", userInfo.Picture)
			user.AvatarURL = userInfo.Picture
		}
		return &user, nil
	}

	if err := db.DB.Where("email = ?", userInfo.Email).First(&user).Error; err == nil {
		user.GoogleID = userInfo.ID
		if userInfo.Picture != "" {
			user.AvatarURL = userInfo.Picture
		}
		if err := db.DB.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user with Google ID: %w", err)
		}
		return &user, nil
	}

	user = model.User{
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		GoogleID:  userInfo.ID,
		AvatarURL: userInfo.Picture,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}
