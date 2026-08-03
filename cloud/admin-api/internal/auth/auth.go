package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.MapClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func ComparePassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

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

type LoginResult struct {
	Token string     `json:"token"`
	User  *UserInfo  `json:"user"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

func GenerateToken(userID, role, secret string) (string, error) {
	if userID == "" || role == "" || secret == "" {
		return "", errors.New("invalid parameters: user_id, role, and secret are required")
	}

	now := time.Now()
	expiry := now.Add(2 * time.Hour)

	claims := Claims{
		MapClaims: jwt.MapClaims{
			"exp":      expiry.Unix(),
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

func ExtractUserIDFromToken(tokenString, secret string) (string, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func ExtractRoleFromToken(tokenString, secret string) (string, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

func VerifyLogin(method, credential, secret string) (*struct {
	ID   string
	Name string
	Role string
}, error) {
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
