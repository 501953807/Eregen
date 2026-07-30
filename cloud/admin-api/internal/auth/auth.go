package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the custom JWT claims structure with user identity and expiration.
type Claims struct {
	jwt.MapClaims // Embed MapClaims for standard JWT claims (exp, iat, nbf, iss, sub, aud, etc.)
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
}

// GenerateToken creates a new JWT token with the given user ID and role.
func GenerateToken(userID, role, secret string) (string, error) {
	if userID == "" || role == "" || secret == "" {
		return "", errors.New("invalid parameters: user_id, role, and secret are required")
	}

	now := time.Now()
	expiry := now.Add(2 * time.Hour)

	claims := Claims{
		MapClaims: jwt.MapClaims{
			"exp":  expiry.Unix(),
			"IssuedAt": now.Unix(),
			"Issuer":   "eregen_admin_api",
			"sub":      userID,
		},
		UserID: userID,
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT token, returning the claims if valid.
func ValidateToken(tokenString, secret string) (*Claims, error) {
	if tokenString == "" || secret == "" {
		return nil, errors.New("token and secret are required")
	}

	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, errors.New("token must be in format 'Bearer <token>'")
	}

	tokenToValidate := parts[1]
	token, err := jwt.ParseWithClaims(tokenToValidate, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	return claims, nil
}

// ExtractUserIDFromToken extracts the user_id from a valid JWT token.
func ExtractUserIDFromToken(tokenString, secret string) (string, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ExtractRoleFromToken extracts the role from a valid JWT token.
func ExtractRoleFromToken(tokenString, secret string) (string, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

// AdminUser represents a default administrative user for initial setup/testing.
type AdminUser struct {
	Username string
	Password string
	Role     string
}

// DefaultAdminUser is the default admin credentials used for testing and development.
var DefaultAdminUser = AdminUser{
	Username: "admin@example.com",
	Password: "Admin@123",
	Role:     "admin",
}

// LoginResult returns the login response containing JWT token and user info.
type LoginResult struct {
	Token string
	User  *UserInfo
}

// UserInfo contains basic user information returned after login.
type UserInfo struct {
	Username string
	Role     string
}

// VerifyLogin checks if the provided username/password match the default admin user.
// This is a temporary placeholder; in production, this should query the database.
func VerifyLogin(username, password string) bool {
	return username == DefaultAdminUser.Username && password == DefaultAdminUser.Password
}
