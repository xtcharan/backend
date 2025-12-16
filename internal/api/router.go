package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/college-event-backend/internal/api/handlers"
	"github.com/yourusername/college-event-backend/internal/api/middleware"
	"github.com/yourusername/college-event-backend/internal/services/auth"
	"github.com/yourusername/college-event-backend/pkg/database"
)

type Router struct {
	engine       *gin.Engine
	db           *database.DB
	authService  *auth.Service
	corsOrigins  string
}

func NewRouter(db *database.DB, authService *auth.Service, corsOrigins string) *Router {
	return &Router{
		engine:      gin.Default(),
		db:          db,
		authService: authService,
		corsOrigins: corsOrigins,
	}
}

func (r *Router) Setup() *gin.Engine {
	// Apply CORS middleware
	r.engine.Use(middleware.CORSMiddleware(r.corsOrigins))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(r.db, r.authService)
	eventHandler := handlers.NewEventHandler(r.db)

	// Health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "college-events-api",
		})
	})

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Public event routes (read-only)
		v1.GET("/events", eventHandler.ListEvents)
		v1.GET("/events/:id", eventHandler.GetEvent)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(r.authService))
		{
			// User profile
			protected.GET("/profile", authHandler.GetProfile)

			// Event registration (students can register for events)
			// protected.POST("/events/:id/register", eventHandler.RegisterForEvent)
			// protected.DELETE("/events/:id/register", eventHandler.UnregisterFromEvent)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(r.authService))
		admin.Use(middleware.AdminMiddleware())
		{
			// Event management
			admin.POST("/events", eventHandler.CreateEvent)
			admin.PUT("/events/:id", eventHandler.UpdateEvent)
			admin.DELETE("/events/:id", eventHandler.DeleteEvent)

			// Club management (to be implemented)
			// admin.POST("/clubs", clubHandler.CreateClub)
			// admin.PUT("/clubs/:id", clubHandler.UpdateClub)
			// admin.DELETE("/clubs/:id", clubHandler.DeleteClub)

			// User management (to be implemented)
			// admin.GET("/users", userHandler.ListUsers)
			// admin.PUT("/users/:id/role", userHandler.UpdateUserRole)
		}
	}

	return r.engine
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
