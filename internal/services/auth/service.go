package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/yourusername/college-event-backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

// Claims represents JWT claims
type Claims struct {
	UserID uuid.UUID       `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Service handles authentication logic
type Service struct {
	jwtSecret              []byte
	jwtExpiryHours         int
	refreshTokenExpiryDays int
}

// NewService creates a new auth service
func NewService(jwtSecret string, jwtExpiryHours int, refreshTokenExpiryDays int) *Service {
	return &Service{
		jwtSecret:              []byte(jwtSecret),
		jwtExpiryHours:         jwtExpiryHours,
		refreshTokenExpiryDays: refreshTokenExpiryDays,
	}
}

// HashPassword generates a bcrypt hash of the password
func (s *Service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares password with hash
func (s *Service) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAccessToken generates a JWT access token
func (s *Service) GenerateAccessToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Hour * time.Duration(s.jwtExpiryHours))

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken() (string, time.Time, error) {
	tokenID := uuid.New()
	expiresAt := time.Now().Add(time.Hour * 24 * time.Duration(s.refreshTokenExpiryDays))
	return tokenID.String(), expiresAt, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// IsAdmin checks if user has admin role
func IsAdmin(role models.UserRole) bool {
	return role == models.RoleAdmin
}

// IsFaculty checks if user has faculty role
func IsFaculty(role models.UserRole) bool {
	return role == models.RoleFaculty
}

// IsAdminOrFaculty checks if user has admin or faculty role
func IsAdminOrFaculty(role models.UserRole) bool {
	return IsAdmin(role) || IsFaculty(role)
}
