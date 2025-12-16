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
	deptHandler := &handlers.DepartmentHandler{DB: r.db.DB}
	clubHandler := &handlers.ClubHandler{DB: r.db.DB}

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

		// ====================================================================
		// PUBLIC ROUTES - Departments & Clubs
		// ====================================================================
		
		// Departments
		v1.GET("/departments", deptHandler.GetDepartments)
		v1.GET("/departments/:id", deptHandler.GetDepartment)
		v1.GET("/departments/:id/clubs", deptHandler.GetDepartmentClubs)

		// Clubs
		v1.GET("/clubs", clubHandler.GetClubs)
		v1.GET("/clubs/:id", clubHandler.GetClub)
		v1.GET("/clubs/:id/members", clubHandler.GetClubMembers)
		v1.GET("/clubs/:id/events", clubHandler.GetClubEvents)
		v1.GET("/clubs/:id/announcements", clubHandler.GetClubAnnouncements)
		v1.GET("/clubs/:id/awards", clubHandler.GetClubAwards)

		// Events
		v1.GET("/events", eventHandler.ListEvents)
		v1.GET("/events/:id", eventHandler.GetEvent)

		// ====================================================================
		// PROTECTED ROUTES (Authenticated Users)
		// ====================================================================
		
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(r.authService))
		{
			// User profile
			protected.GET("/profile", authHandler.GetProfile)

			// Club announcements (create/update/delete by club admins)
			protected.POST("/clubs/:id/announcements", clubHandler.CreateClubAnnouncement)
			protected.PUT("/clubs/:id/announcements/:announcement_id", clubHandler.UpdateClubAnnouncement)
			protected.DELETE("/clubs/:id/announcements/:announcement_id", clubHandler.DeleteClubAnnouncement)

			// Club members (add/update/remove by club admins)
			protected.POST("/clubs/:id/members", clubHandler.AddClubMember)
			protected.PUT("/clubs/:id/members/:user_id", clubHandler.UpdateClubMember)
			protected.DELETE("/clubs/:id/members/:user_id", clubHandler.RemoveClubMember)

			// Club awards (add by club admins)
			protected.POST("/clubs/:id/awards", clubHandler.CreateClubAward)
		}

		// ====================================================================
		// ADMIN ROUTES (System Administrators)
		// ====================================================================
		
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(r.authService))
		admin.Use(middleware.AdminMiddleware())
		{
			// Department management
			admin.POST("/departments", deptHandler.CreateDepartment)
			admin.PUT("/departments/:id", deptHandler.UpdateDepartment)
			admin.DELETE("/departments/:id", deptHandler.DeleteDepartment)

			// Club management
			admin.POST("/clubs", clubHandler.CreateClub)
			admin.PUT("/clubs/:id", clubHandler.UpdateClub)
			admin.DELETE("/clubs/:id", clubHandler.DeleteClub)

			// Event management
			admin.POST("/events", eventHandler.CreateEvent)
			admin.PUT("/events/:id", eventHandler.UpdateEvent)
			admin.DELETE("/events/:id", eventHandler.DeleteEvent)
		}
	}

	return r.engine
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
