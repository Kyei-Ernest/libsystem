package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/user-service/handlers"
	"github.com/Kyei-Ernest/libsystem/services/user-service/repository"
	"github.com/Kyei-Ernest/libsystem/services/user-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/database"
	sharedRedis "github.com/Kyei-Ernest/libsystem/shared/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Kyei-Ernest/libsystem/services/user-service/docs"
)

// @title           User Service API
// @version         1.0
// @description     API for User Management and Authentication
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8086
// @BasePath        /api/v1
func main() {
	// Load .env file (optional - won't fail if missing)
	_ = godotenv.Load("../../.env")

	// Load configuration from environment
	port := getEnv("PORT", "8080")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "libsystem")
	dbPassword := getEnv("DB_PASSWORD", "libsystem")
	dbName := getEnv("DB_NAME", "libsystem")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production-min-32-chars")

	// Initialize database connection
	dbConfig := &database.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	dbConn, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Note: AutoMigrate is disabled because all tables are created by SQL migrations
	// The migrations in migrations/*.up.sql handle all schema creation
	// Uncomment below only if you need GORM to manage schema (not recommended)
	// if err := dbConn.DB.AutoMigrate(&models.User{}); err != nil {
	// 	log.Fatalf("Failed to migrate database: %v", err)
	// }

	log.Println("Database connected successfully")

	// Initialize Redis client (optional - for token blacklisting)
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")

	redisConfig := &sharedRedis.Config{
		Host:     redisHost,
		Port:     redisPort,
		Password: redisPassword,
		DB:       0,
	}

	var blacklistService service.TokenBlacklistService
	redisClient, err := sharedRedis.NewClient(redisConfig)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v (token blacklisting disabled)", err)
		blacklistService = nil // Service will work without Redis, just no immediate logout
	} else {
		log.Println("Redis connected successfully")
		blacklistService = service.NewTokenBlacklistService(redisClient)
		defer redisClient.Close()
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbConn.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, blacklistService, jwtSecret)
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userService)
	userHandler := handlers.NewUserHandler(userService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := dbConn.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"service":  "user-service",
			"database": "connected",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		authHandler.RegisterRoutes(v1)
		authMiddleware := handlers.AuthMiddleware(authService)
		userHandler.RegisterRoutes(v1, authMiddleware)
	}

	// Swagger configuration
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("User Service starting on port %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	if err := srv.Close(); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
