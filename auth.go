package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ── Models ───────────────────────────────────────────────────────────────────

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ── Storage ──────────────────────────────────────────────────────────────────

const usersFile = "data/users.json"

var (
	usersMu sync.RWMutex
	jwtSecret []byte
)

func initAuth() {
	// JWT secret: env var or random on each start (dev mode)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	jwtSecret = []byte(secret)

	// Ensure data dir exists
	os.MkdirAll("data", 0755)

	// Seed default admin user if users.json missing
	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		users := []User{{Username: "admin", PasswordHash: string(hash)}}
		data, _ := json.MarshalIndent(users, "", "  ")
		os.WriteFile(usersFile, data, 0600)
	}
}

func loadUsers() ([]User, error) {
	usersMu.RLock()
	defer usersMu.RUnlock()
	data, err := os.ReadFile(usersFile)
	if err != nil {
		return nil, err
	}
	var users []User
	return users, json.Unmarshal(data, &users)
}

func findUser(username string) (*User, error) {
	users, err := loadUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if strings.EqualFold(u.Username, username) {
			return &u, nil
		}
	}
	return nil, nil
}

// ── Token generation ─────────────────────────────────────────────────────────

func generateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// ── Middleware ────────────────────────────────────────────────────────────────

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := validateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz veya süresi dolmuş token"})
			return
		}
		c.Set("username", claims.Username)
		c.Next()
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kullanıcı adı ve şifre gerekli"})
		return
	}

	user, err := findUser(req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz kullanıcı adı veya şifre"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz kullanıcı adı veya şifre"})
		return
	}

	token, err := generateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token oluşturulamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": user.Username,
	})
}
