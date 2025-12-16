package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/college-event-backend/internal/models"
	"github.com/yourusername/college-event-backend/internal/services/auth"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   strPtr("missing authorization header"),
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   strPtr("invalid authorization header format"),
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   strPtr("invalid or expired token"),
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   strPtr("unauthorized"),
			})
			c.Abort()
			return
		}

		userRole, ok := role.(models.UserRole)
		if !ok || !auth.IsAdmin(userRole) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("admin access required"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOrFacultyMiddleware checks if user has admin or faculty role
func AdminOrFacultyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   strPtr("unauthorized"),
			})
			c.Abort()
			return
		}

		userRole, ok := role.(models.UserRole)
		if !ok || !auth.IsAdminOrFaculty(userRole) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   strPtr("admin or faculty access required"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func strPtr(s string) *string {
	return &s
}
