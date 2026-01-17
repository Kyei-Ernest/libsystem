package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/analytics-service/consumer"
	"github.com/Kyei-Ernest/libsystem/services/analytics-service/handlers"
	"github.com/Kyei-Ernest/libsystem/services/analytics-service/models"
	"github.com/Kyei-Ernest/libsystem/services/analytics-service/repository"
	"github.com/Kyei-Ernest/libsystem/shared/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Kyei-Ernest/libsystem/services/analytics-service/docs"
)

// @title           Analytics Service API
// @version         1.0
// @description     API for system analytics and statistics
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8087
// @BasePath        /api/v1/analytics

func main() {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Database
	dbConfig := database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		TimeZone: "UTC",
	}

	conn, err := database.NewConnection(&dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db := conn.DB

	// Auto Migrate
	log.Println("Migrating database...")
	if err := db.AutoMigrate(&models.AnalyticsEvent{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Components
	repo := repository.NewAnalyticsRepository(db)
	handler := handlers.NewAnalyticsHandler(repo)

	// Kafka Config
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"localhost:9093"}
	}

	// Start Consumer
	analConsumer := consumer.NewAnalyticsConsumer(brokers, "analytics-service", repo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := analConsumer.Start(ctx, brokers); err != nil {
		log.Printf("Failed to start consumers: %v", err)
	}

	// Initialize Router
	router := gin.Default()

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API Routes
	api := router.Group("/api/v1/analytics")
	{
		api.GET("/overview", handler.GetOverview)
		api.GET("/documents/popular", handler.GetTopDocuments)
		api.GET("/activity", handler.GetActivity)
	}

	// Swagger configuration
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start Server
	port := os.Getenv("ANALYTICS_SERVICE_PORT")
	if port == "" {
		port = "8087"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Analytics Service starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
