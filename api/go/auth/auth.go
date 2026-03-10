package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token   string     `json:"token"`
	User    model.User `json:"user"`
	Message string     `json:"message"`
}

var googleOAuthConfig *oauth2.Config

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

	var existingUser model.User
	if err := db.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := model.User{
		Name:         req.Name,
		Email:        req.Email,
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

	var user model.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
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

	url := config.AuthCodeURL("random-state-string", oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func GoogleCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code parameter"})
		return
	}

	config := getGoogleOAuthConfig()

	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	client := config.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	if !userInfo.VerifiedEmail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google email not verified"})
		return
	}

	user, err := findOrCreateGoogleUser(userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	jwtToken, err := GenerateJWTToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/google/callback?token=%s", frontendURL, jwtToken))
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
