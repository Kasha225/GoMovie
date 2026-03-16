package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
)

func InitAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-pls-change"
	}
	jwtSecret = []byte(secret)
	accessTTL = 15 * time.Minute
	refreshTTL = 7 * time.Hour
}
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func CreateAccessToken(userID int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"iat": now.Unix(),
		"exp": now.Add(accessTTL).Unix(),
		"typ": "access",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(jwtSecret)
}

func CreateRefreshToken(userID int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"iat": now.Unix(),
		"exp": now.Add(refreshTTL).Unix(),
		"typ": "refresh",
		"rnd": fmt.Sprintf("%d", now.UnixNano()), // уникальность
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(jwtSecret)
}
