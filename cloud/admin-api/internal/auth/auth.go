package auth

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// HashPassword generates a bcrypt hash from the given password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ComparePassword checks if the given password matches the stored hash.
func ComparePassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// DefaultUsers is the built-in default users for development (passwords are plaintext for this list; see seedDatabase for bcrypt hashes)
var DefaultUsers = []struct {
	Username string
	Password string
	Phone    string
	OTP      string
	Role     string
	Name     string
	ID       string
}{
	{Username: "admin@eregen.com", Password: "Admin@123", Role: "admin", Name: "系统管理员", ID: "usr-admin"},
	{Username: "family@eregen.com", Password: "Family@123", Phone: "13800000002", OTP: "123456", Role: "family", Name: "张伟", ID: "usr-fam1"},
	{Username: "operator@eregen.com", Password: "Op@1234", Phone: "13900000003", OTP: "123456", Role: "operator", Name: "李护士", ID: "usr-op1"},
}

// LoginResult represents the result of a successful login
type LoginResult struct {
	Token string     `json:"token"`
	User  *UserInfo  `json:"user"`
}

// UserInfo contains basic user info returned after login
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

// GenerateToken creates a JWT token for the given user
func GenerateToken(userID, role, secret string, log *zap.Logger) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(2 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

// VerifyLogin checks if the provided credentials match any default user.
// method is "email" (uses Username) or "phone" (uses Phone/OTP).
// Phone credentials are normalized by stripping +86/86 prefix.
func VerifyLogin(method, credential, secret string) (*struct {
	ID   string
	Name string
	Role string
}, error) {
	// Normalize phone: strip +86 or 86 prefix for comparison
	normalized := strings.TrimPrefix(credential, "+86")
	normalized = strings.TrimPrefix(normalized, "86")
	for _, u := range DefaultUsers {
		if method == "email" && u.Username == credential && u.Password == secret {
			return &struct{ ID, Name, Role string }{u.ID, u.Name, u.Role}, nil
		}
		if method == "phone" && (u.Phone == credential || u.Phone == normalized) && u.OTP == secret {
			return &struct{ ID, Name, Role string }{u.ID, u.Name, u.Role}, nil
		}
	}
	return nil, fmt.Errorf("invalid credentials")
}
