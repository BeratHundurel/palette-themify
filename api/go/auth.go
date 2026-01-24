package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
	Token   string `json:"token"`
	User    User   `json:"user"`
	Message string `json:"message"`
}

type DemoLoginRequest struct {
	// No fields needed for demo login
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWTToken(user User) (string, error) {
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

func createDemoUserIfNotExists() (*User, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	demoEmail := "demo@imagepalette.com"
	var demoUser User

	if err := DB.Where("email = ?", demoEmail).First(&demoUser).Error; err == nil {
		return &demoUser, nil
	}

	hashedPassword, err := hashPassword("demopassword123")
	if err != nil {
		return nil, fmt.Errorf("failed to hash demo password: %w", err)
	}

	demoUser = User{
		Name:         "Demo User",
		Email:        demoEmail,
		PasswordHash: hashedPassword,
	}

	if err := DB.Create(&demoUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create demo user: %w", err)
	}

	if err := createSamplePalettes(demoUser.ID); err != nil {
		fmt.Printf("Warning: Failed to create sample palettes for demo user: %v\n", err)
	}

	return &demoUser, nil
}

func createSamplePalettes(userID uint) error {
	samplePalettes := []struct {
		Name     string
		JsonData string
	}{
		{
			Name:     "Ocean Sunset",
			JsonData: `[{"hex":"#FF6B35"},{"hex":"#F7931E"},{"hex":"#FFD23F"},{"hex":"#06FFA5"},{"hex":"#4ECDC4"},{"hex":"#1B9AAA"},{"hex":"#EF476F"},{"hex":"#FFC43D"}]`,
		},
		{
			Name:     "Forest Vibes",
			JsonData: `[{"hex":"#2D5016"},{"hex":"#61A05D"},{"hex":"#8FBC8F"},{"hex":"#B5D6AA"},{"hex":"#D6EAD0"},{"hex":"#4A6741"},{"hex":"#7BA05B"},{"hex":"#A8D8A8"}]`,
		},
		{
			Name:     "Retro Gaming",
			JsonData: `[{"hex":"#FF0040"},{"hex":"#FF8000"},{"hex":"#FFFF00"},{"hex":"#00FF00"},{"hex":"#0080FF"},{"hex":"#8000FF"},{"hex":"#FF00FF"},{"hex":"#00FFFF"}]`,
		},
		{
			Name:     "Purple Dreams",
			JsonData: `[{"hex":"#4A154B"},{"hex":"#7B2982"},{"hex":"#A663CC"},{"hex":"#D4B2F7"},{"hex":"#F0E6FF"},{"hex":"#6B2C91"},{"hex":"#9B59B6"},{"hex":"#BB8FCE"}]`,
		},
		{
			Name:     "Warm Autumn",
			JsonData: `[{"hex":"#8B4513"},{"hex":"#D2691E"},{"hex":"#FF8C00"},{"hex":"#FFA500"},{"hex":"#FFD700"},{"hex":"#CD853F"},{"hex":"#DEB887"},{"hex":"#F4A460"}]`,
		},
		{
			Name:     "Grayscale Classic",
			JsonData: `[{"hex":"#E3E7E7"},{"hex":"#B5B5AF"},{"hex":"#CCCECB"},{"hex":"#959590"},{"hex":"#6A6C6B"},{"hex":"#343638"},{"hex":"#BA875C"},{"hex":"#85593C"}]`,
		},
	}

	for _, palette := range samplePalettes {
		dbPalette := Palette{
			UserID:   &userID,
			Name:     palette.Name,
			JsonData: palette.JsonData,
			IsSystem: true,
		}

		if err := DB.Create(&dbPalette).Error; err != nil {
			return fmt.Errorf("failed to create palette %s: %w", palette.Name, err)
		}
	}

	return nil
}

func validateJWTToken(tokenString string) (*Claims, error) {
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

func authMiddleware() gin.HandlerFunc {
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

		claims, err := validateJWTToken(bearerToken[1])
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

func getCurrentUser(c *gin.Context) (uint, error) {
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

func registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser User
	if err := DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := generateJWTToken(user)
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

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !checkPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := generateJWTToken(user)
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

func getMeHandler(c *gin.Context) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func changePasswordHandler(c *gin.Context) {
	userID, err := getCurrentUser(c)
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

	var user User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !checkPasswordHash(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	if err := DB.Model(&user).Update("password_hash", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func demoLoginHandler(c *gin.Context) {
	demoUser, err := createDemoUserIfNotExists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create demo user"})
		return
	}

	token, err := generateJWTToken(*demoUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	demoUser.PasswordHash = ""
	c.JSON(http.StatusOK, AuthResponse{
		Token:   token,
		User:    *demoUser,
		Message: "Demo login successful",
	})
}
