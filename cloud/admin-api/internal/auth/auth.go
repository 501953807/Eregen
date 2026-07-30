package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// DefaultAdminUser is the built-in default admin user for development
var DefaultAdminUser = struct {
	Username string
	Password string // In production, use hashed password
	Role     string
}{
	Username: "admin",
	Password: "Admin@123", // Pre-stated password for quick setup
	Role:     "super_admin",
}

// LoginResult represents the result of a successful login
type LoginResult struct {
	Token  string      `json:"token"`
	User   *UserInfo   `json:"user"`
}

// UserInfo contains basic user info returned after login
type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// GenerateToken creates a JWT token for the given user
func GenerateToken(username string, role string, secret string, log *zap.Logger) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

// VerifyLogin checks if the provided credentials match the default admin user
func VerifyLogin(username, password string) bool {
	return username == DefaultAdminUser.Username && password == DefaultAdminUser.Password
}
