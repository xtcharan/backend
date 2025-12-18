package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/college-event-backend/internal/api/handlers"
	"github.com/yourusername/college-event-backend/internal/api/middleware"
	"github.com/yourusername/college-event-backend/internal/services/auth"
	"github.com/yourusername/college-event-backend/internal/storage"
	"github.com/yourusername/college-event-backend/pkg/database"
)

type Router struct {
	engine      *gin.Engine
	db          *database.DB
	authService *auth.Service
	storage     storage.StorageService
	corsOrigins string
}

func NewRouter(db *database.DB, authService *auth.Service, storageService storage.StorageService, corsOrigins string) *Router {
	return &Router{
		engine:      gin.Default(),
		db:          db,
		authService: authService,
		storage:     storageService,
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
	scheduleHandler := handlers.NewScheduleHandler(r.db)
	uploadHandler := handlers.NewUploadHandler(r.storage)
	houseHandler := handlers.NewHouseHandler(r.db.DB)
	postsHandler := handlers.NewPostsHandler(r.db.DB)
	storiesHandler := handlers.NewStoriesHandler(r.db.DB)
	paymentHandler := handlers.NewPaymentHandler(r.db)

	// Health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "college-events-api",
		})
	})

	// Serve static files for local storage (development)
	r.engine.Static("/uploads", "./uploads")

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleAuth)
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

		// Schedules (public GET - returns official schedules, personal schedules if authenticated)
		v1.GET("/schedules", middleware.OptionalAuthMiddleware(r.authService), scheduleHandler.ListSchedules)
		v1.GET("/schedules/:id", middleware.OptionalAuthMiddleware(r.authService), scheduleHandler.GetSchedule)

		// Houses (public)
		v1.GET("/houses", houseHandler.GetHouses)
		v1.GET("/houses/:id", houseHandler.GetHouse)
		v1.GET("/houses/:id/announcements", middleware.OptionalAuthMiddleware(r.authService), houseHandler.GetAnnouncements)
		v1.GET("/houses/:id/events", middleware.OptionalAuthMiddleware(r.authService), houseHandler.GetHouseEvents)
		v1.GET("/announcements/:id/comments", houseHandler.GetComments)

		// Posts (public read, authenticated for interactions)
		v1.GET("/posts", middleware.OptionalAuthMiddleware(r.authService), postsHandler.ListPosts)
		v1.GET("/posts/:id", middleware.OptionalAuthMiddleware(r.authService), postsHandler.GetPost)
		v1.POST("/posts/:id/view", postsHandler.TrackView) // Can be anonymous

		// Stories (public read, authenticated for interactions)
		v1.GET("/stories", middleware.OptionalAuthMiddleware(r.authService), storiesHandler.ListStories)

		// ====================================================================
		// PROTECTED ROUTES (Authenticated Users)
		// ====================================================================

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(r.authService))
		{
			// User profile
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile", authHandler.UpdateProfile)

			// ================================================================
			// PAYMENT ROUTES - Razorpay Integration
			// ================================================================
			payments := protected.Group("/payments")
			{
				payments.POST("/create-order", paymentHandler.CreateOrder)
				payments.POST("/verify", paymentHandler.VerifyPayment)
				payments.GET("/status/:event_id", paymentHandler.GetPaymentStatus)
			}

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

			// Schedule management (users can create/edit/delete their own personal schedules)
			protected.POST("/schedules", scheduleHandler.CreateSchedule)
			protected.PUT("/schedules/:id", scheduleHandler.UpdateSchedule)
			protected.DELETE("/schedules/:id", scheduleHandler.DeleteSchedule)

			// House interactions (authenticated users)
			protected.POST("/houses/:id/roles", houseHandler.AddHouseRole)
			protected.DELETE("/houses/:id/roles/:role_id", houseHandler.RemoveHouseRole)
			protected.POST("/announcements/:id/like", houseHandler.LikeAnnouncement)
			protected.POST("/announcements/:id/comments", houseHandler.AddComment)
			protected.POST("/house-events/:event_id/enroll", houseHandler.EnrollInEvent)
			protected.DELETE("/house-events/:event_id/enroll", houseHandler.UnenrollFromEvent)

			// Post interactions (authenticated users)
			protected.POST("/posts/:id/like", postsHandler.ToggleLike)
			protected.POST("/posts/:id/comment", postsHandler.AddComment)
			protected.DELETE("/posts/:id/comments/:comment_id", postsHandler.DeleteComment)
			protected.POST("/posts/:id/share", postsHandler.TrackShare)

			// Story interactions (authenticated users)
			protected.POST("/stories/:id/like", storiesHandler.ToggleLike)
			protected.POST("/stories/:id/view", storiesHandler.TrackView)
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

			// Image upload (optimized & stored to GCS/local)
			admin.POST("/upload", uploadHandler.UploadImage)

			// House management
			admin.POST("/houses", houseHandler.CreateHouse)
			admin.PUT("/houses/:id", houseHandler.UpdateHouse)
			admin.DELETE("/houses/:id", houseHandler.DeleteHouse)
			admin.POST("/houses/:id/announcements", houseHandler.CreateAnnouncement)
			admin.POST("/houses/:id/events", houseHandler.CreateHouseEvent)

			// Posts management (admin/faculty only)
			admin.POST("/posts", postsHandler.CreatePost)
			admin.PUT("/posts/:id", postsHandler.UpdatePost)
			admin.DELETE("/posts/:id", postsHandler.DeletePost)          // Soft delete
			admin.DELETE("/posts/:id/hard", postsHandler.HardDeletePost) // Permanent delete

			// Stories management (admin/faculty only)
			admin.POST("/stories", storiesHandler.CreateStory)
			admin.DELETE("/stories/:id/hard", storiesHandler.HardDeleteStory) // Permanent delete
		}
	}

	return r.engine
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
