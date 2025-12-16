package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/college-event-backend/internal/models"
	"github.com/yourusername/college-event-backend/internal/services/auth"
	"github.com/yourusername/college-event-backend/pkg/database"
)

type AuthHandler struct {
	db          *database.DB
	authService *auth.Service
}

func NewAuthHandler(db *database.DB, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		db:          db,
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid request body"),
		})
		return
	}

	// Check if user already exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)", req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("database error"),
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error:   strPtr("user already exists"),
		})
		return
	}

	// Hash password
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to hash password"),
		})
		return
	}

	// Create user
	var user models.User
	err = h.db.QueryRow(`
		INSERT INTO users (email, password_hash, full_name, role, department, year)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, full_name, role, department, year, created_at, updated_at
	`, req.Email, passwordHash, req.FullName, models.RoleStudent, req.Department, req.Year).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role,
		&user.Department, &user.Year, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to create user"),
		})
		return
	}

	// Generate tokens
	accessToken, err := h.authService.GenerateAccessToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to generate token"),
		})
		return
	}

	refreshToken, expiresAt, err := h.authService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to generate refresh token"),
		})
		return
	}

	// Store refresh token
	_, err = h.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, refreshToken, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to store refresh token"),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "user registered successfully",
		Data: models.LoginResponse{
			User:         user,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid request body"),
		})
		return
	}

	// Get user by email
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, email, password_hash, full_name, role, avatar_url, department, year, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role,
		&user.AvatarURL, &user.Department, &user.Year, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid credentials"),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("database error"),
		})
		return
	}

	// Check password
	if !h.authService.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid credentials"),
		})
		return
	}

	// Generate tokens
	accessToken, err := h.authService.GenerateAccessToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to generate token"),
		})
		return
	}

	refreshToken, expiresAt, err := h.authService.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to generate refresh token"),
		})
		return
	}

	// Store refresh token
	_, err = h.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, refreshToken, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to store refresh token"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "login successful",
		Data: models.LoginResponse{
			User:         user,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, email, full_name, role, avatar_url, department, year, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID.(uuid.UUID)).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role,
		&user.AvatarURL, &user.Department, &user.Year, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to fetch profile"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

func strPtr(s string) *string {
	return &s
}
