package main

import (
	"log"
	"os"

	"github.com/Kyei-Ernest/libsystem/services/search-service/handlers"
	"github.com/Kyei-Ernest/libsystem/services/search-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/elasticsearch"
	"github.com/Kyei-Ernest/libsystem/shared/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Kyei-Ernest/libsystem/services/search-service/docs"
)

// @title           Search Service API
// @version         1.0
// @description     API for searching documents
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8084
// @BasePath        /api/v1/search

func main() {
	// Config
	esAddress := getEnv("ELASTICSEARCH_URL", "http://localhost:9200")
	port := getEnv("PORT", "8084")

	log.Println("Search Service Starting...")

	// Initialize Elasticsearch Client
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esAddress},
	})
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Initialize Service & Logic
	searchSvc := service.NewSearchService(esClient)
	searchHandler := handlers.NewSearchHandler(searchSvc)

	// Initialize router
	router := gin.Default()

	// Add Prometheus metrics middleware
	router.Use(metrics.PrometheusMiddleware())

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "search-service"})
	})

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API Routes
	api := router.Group("/api/v1/search")
	{
		api.GET("", searchHandler.Search)
		api.GET("/advanced", searchHandler.AdvancedSearch)
	}

	// Swagger configuration
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
