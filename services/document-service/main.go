package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/document-service/handlers"
	"github.com/Kyei-Ernest/libsystem/services/document-service/middleware"
	"github.com/Kyei-Ernest/libsystem/services/document-service/repository"
	"github.com/Kyei-Ernest/libsystem/services/document-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/database"
	"github.com/Kyei-Ernest/libsystem/shared/jobs"
	"github.com/Kyei-Ernest/libsystem/shared/kafka"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/security"
	"github.com/Kyei-Ernest/libsystem/shared/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Kyei-Ernest/libsystem/services/document-service/docs"
)

// @title           Document Service API
// @version         1.0
// @description     API for managing documents and file storage
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8081
// @BasePath        /api/v1

func main() {
	// Load .env file (optional - won't fail if missing)
	_ = godotenv.Load("../../.env")

	// Load configuration from environment
	port := getEnv("PORT", "8081")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "libsystem")
	dbPassword := getEnv("DB_PASSWORD", "libsystem")
	dbName := getEnv("DB_NAME", "libsystem")

	// Kafka Config
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9093"), ",")

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

	log.Println("Database connected successfully")

	// Initialize MinIO storage client
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin123")
	minioUseSSL := getEnv("MINIO_USE_SSL", "false") == "true"
	minioBucket := getEnv("MINIO_BUCKET_DOCUMENTS", "documents")

	minioConfig := &storage.MinIOConfig{
		Endpoint:        minioEndpoint,
		AccessKeyID:     minioAccessKey,
		SecretAccessKey: minioSecretKey,
		UseSSL:          minioUseSSL,
		BucketName:      minioBucket,
		Region:          "us-east-1",
	}

	var storageClient *storage.MinIOClient
	minioClient, err := storage.NewMinIOClient(minioConfig)
	if err != nil {
		log.Printf("Warning: Failed to connect to MinIO: %v (file storage disabled)", err)
		storageClient = nil
	} else {
		log.Println("MinIO connected successfully")
		storageClient = minioClient
		defer storageClient.Close()
	}

	// Initialize Kafka Producer
	log.Printf("Connecting to Kafka brokers: %v", kafkaBrokers)
	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: kafkaBrokers,
		Topic:   "", // No default topic, we specify per message
	})
	defer producer.Close()

	// Initialize Virus Scanner (optional - will disable if ClamAV not available)
	clamavAddr := getEnv("CLAMAV_ADDR", "tcp://localhost:3310")
	var virusScanner *security.VirusScanner
	scanner, err := security.NewVirusScanner(clamavAddr)
	if err != nil {
		log.Printf("Warning: Virus scanning disabled - ClamAV not available: %v", err)
		virusScanner = nil
	} else {
		log.Println("Virus scanner initialized successfully")
		virusScanner = scanner
	}

	// Initialize services
	documentRepo := repository.NewDocumentRepository(dbConn.DB)
	collectionRepo := repository.NewCollectionRepository(dbConn.DB)
	permissionRepo := repository.NewPermissionRepository(dbConn.DB)
	fileService := service.NewFileService()
	documentService := service.NewDocumentService(documentRepo, collectionRepo, fileService, storageClient, producer, virusScanner)
	permissionService := service.NewPermissionService(permissionRepo, documentRepo, collectionRepo)

	// Initialize job tracker
	jobTracker := jobs.NewJobTracker()

	// Initialize handlers
	documentHandler := handlers.NewDocumentHandler(documentService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	batchHandler := handlers.NewBatchHandler(documentService, jobTracker)

	// Initialize middleware
	permissionChecker := middleware.NewPermissionChecker(permissionService)

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
			"service":  "document-service",
			"database": "connected",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		optionalAuth := optionalAuthMiddleware()
		requiredAuth := requiredAuthMiddleware()
		documentHandler.RegisterRoutes(v1, optionalAuth, requiredAuth, permissionHandler, permissionChecker)

		// Batch operations routes
		batch := v1.Group("/documents/batch")
		{
			batch.POST("/upload", requiredAuth, batchHandler.BulkUpload)
			batch.PATCH("/metadata", requiredAuth, batchHandler.BulkUpdateMetadata)
			batch.DELETE("/delete", requiredAuth, batchHandler.BulkDelete)
		}

		// Job tracking routes
		jobs := v1.Group("/jobs")
		{
			jobs.GET("", requiredAuth, batchHandler.ListJobs)
			jobs.GET("/:jobID", requiredAuth, batchHandler.GetJobStatus)
		}
	}

	// Swagger configuration
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second, // Longer timeout for file uploads
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Document Service starting on port %s...\n", port)
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

// optionalAuthMiddleware extracts user ID if token is present
func optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString != "" && len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			// Token validation would go here
			// c.Set("user_id", userID)
		}
		c.Next()
	}
}

// requiredAuthMiddleware requires authentication
func requiredAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for Service Secret (Internal Auth)
		serviceSecret := getEnv("SERVICE_SECRET", "internal-secret-key")
		if secret := c.GetHeader("X-Service-Secret"); secret != "" {
			if secret == serviceSecret {
				// Internal service call, verify as system (Nil UUID)
				log.Println("DEBUG: Service secret matched, granting system access")
				c.Set("user_id", uuid.Nil)
				c.Next()
				return
			}
			log.Printf("DEBUG: Service secret mismatch. Expected: '%s', Got: '%s'", serviceSecret, secret)
			// If provided but wrong, fail
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid service secret"})
			c.Abort()
			return
		}

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

		// Validate JWT token and extract user ID
		jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production-min-32-chars")
		userID, role, err := validateTokenAndGetUser(tokenString, jwtSecret)
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

		// Set actual user ID and Role from token
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// validateTokenAndGetUserID validates JWT and extracts user ID and Role
func validateTokenAndGetUser(tokenString, jwtSecret string) (uuid.UUID, models.UserRole, error) {
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
		return claims.UserID, claims.Role, nil
	}

	return uuid.Nil, "", fmt.Errorf("invalid token claims")
}
