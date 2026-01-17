package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/collection-service/handlers"
	"github.com/Kyei-Ernest/libsystem/services/collection-service/repository"
	"github.com/Kyei-Ernest/libsystem/services/collection-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/database"
	"github.com/Kyei-Ernest/libsystem/shared/metrics"
	"github.com/Kyei-Ernest/libsystem/shared/security"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Kyei-Ernest/libsystem/services/collection-service/docs"
)

// @title           Collection Service API
// @version         1.0
// @description     API for managing document collections
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8082
// @BasePath        /api/v1

func main() {
	// Load .env file (optional - won't fail if missing)
	_ = godotenv.Load("../../.env")

	// Load configuration from environment
	port := getEnv("PORT", "8082")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "libsystem")
	dbPassword := getEnv("DB_PASSWORD", "libsystem")
	dbName := getEnv("DB_NAME", "libsystem")

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
	// if err := dbConn.DB.AutoMigrate(&models.Collection{}); err != nil {
	// 	log.Fatalf("Failed to migrate database: %v", err)
	// }

	log.Println("Database connected successfully")

	// Initialize repositories
	collectionRepo := repository.NewCollectionRepository(dbConn.DB)

	// Initialize services
	collectionService := service.NewCollectionService(collectionRepo)

	// Initialize handlers
	collectionHandler := handlers.NewCollectionHandler(collectionService)

	// Initialize router
	router := gin.Default()

	// Add Prometheus metrics middleware
	router.Use(metrics.PrometheusMiddleware())

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
			"service":  "collection-service",
			"database": "connected",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Optional auth middleware for public endpoints
		optionalAuth := optionalAuthMiddleware()
		// Required auth middleware for protected endpoints
		requiredAuth := requiredAuthMiddleware()

		collectionHandler.RegisterRoutes(v1, optionalAuth, requiredAuth)
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
		log.Printf("Collection Service starting on port %s...\n", port)
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

// optionalAuthMiddleware extracts user ID if token is present but doesn't require it
func optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString != "" && len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			// Token validation would go here (call user service or validate JWT)
			// For now, we'll just extract a dummy user ID
			// In production, this should validate the token
			// userID := validateTokenAndGetUserID(tokenString[7:])
			// if userID != nil {
			//     c.Set("user_id", *userID)
			// }
		}
		c.Next()
	}
}

// requiredAuthMiddleware requires authentication
func requiredAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "No token provided",
				},
			})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Validate JWT token and extract user ID and role
		jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
		userID, role, err := validateTokenAndGetUserID(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid token: " + err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Set user ID and role in context
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// validateTokenAndGetUserID validates JWT and extracts user ID and role
func validateTokenAndGetUserID(tokenString, jwtSecret string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &security.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	if claims, ok := token.Claims.(*security.TokenClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, "", err
		}
		return userID, string(claims.Role), nil
	}

	return uuid.Nil, "", fmt.Errorf("invalid token")
}
