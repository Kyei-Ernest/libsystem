package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	ratelimit "github.com/Kyei-Ernest/libsystem/services/api-gateway/middleware"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/security"
)

// Config holds server configuration
type Config struct {
	Port              string
	Environment       string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
	MaxRequestSize    int64
	RateLimitRequests int
	RateLimitWindow   time.Duration
	JWTSecret         string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
}

// Server represents the API Gateway server
type Server struct {
	router      *gin.Engine
	config      *Config
	logger      *zap.Logger
	redisClient *redis.Client
	rateLimiter *ratelimit.Limiter
	server      *http.Server
}

func NewServer(cfg *Config, logger *zap.Logger) *Server {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Initialize Redis client
	var redisClient *redis.Client
	var rateLimiter *ratelimit.Limiter

	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})

		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis connection failed, rate limiting will be disabled",
				zap.Error(err),
				zap.String("redis_addr", cfg.RedisAddr))
			redisClient = nil
		} else {
			logger.Info("Redis connected successfully", zap.String("addr", cfg.RedisAddr))

			// Initialize rate limiter with default config
			rateLimiter = ratelimit.NewLimiter(redisClient, ratelimit.DefaultConfig())
		}
	} else {
		logger.Warn("Redis not configured, rate limiting will be disabled")
	}

	s := &Server{
		config:      cfg,
		router:      router,
		logger:      logger,
		redisClient: redisClient,
		rateLimiter: rateLimiter,
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.server = &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return s
}

func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Structured logging middleware
	s.router.Use(s.loggingMiddleware())

	// Request ID middleware
	s.router.Use(s.requestIDMiddleware())

	// CORS middleware
	s.router.Use(s.corsMiddleware())

	// Rate limiting middleware
	s.router.Use(s.rateLimitMiddleware())

	// Security headers
	s.router.Use(s.securityHeadersMiddleware())

	// Request size limit
	s.router.Use(s.maxRequestSizeMiddleware())
}

func (s *Server) setupRoutes() {
	// Health check endpoints
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/ready", s.readinessCheck)

	// Metrics endpoint for Prometheus
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API version 1
	v1 := s.router.Group("/api/v1")
	{
		// Document routes
		documents := v1.Group("/documents")
		{
			documents.POST("", s.uploadDocument)
			documents.GET("/:id", s.getDocument)
			documents.PUT("/:id", s.updateDocument)
			documents.DELETE("/:id", s.deleteDocument)
			documents.GET("", s.listDocuments)
			documents.GET("/:id/download", s.downloadDocument)
			documents.GET("/:id/view", s.viewDocument)
			documents.GET("/:id/thumbnail", s.getThumbnail)
		}

		// Search routes
		search := v1.Group("/search")
		{
			search.GET("", s.search)
			search.POST("/advanced", s.advancedSearch)
		}

		// Collection routes
		collections := v1.Group("/collections")
		{
			collections.POST("", s.createCollection)
			collections.GET("/:id", s.getCollection)
			collections.PUT("/:id", s.updateCollection)
			collections.DELETE("/:id", s.deleteCollection)
			collections.GET("", s.listCollections)
			collections.GET("/:id/documents", s.getCollectionDocuments)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.POST("/register", s.registerUser)
			users.POST("/login", s.loginUser)
			users.GET("/profile", s.authMiddleware(), s.getUserProfile)
			users.PUT("/profile", s.authMiddleware(), s.updateUserProfile)
			users.DELETE("/:id", s.authMiddleware(), s.deleteUser)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(s.authMiddleware(), s.adminMiddleware())
		{
			admin.GET("/stats", s.getSystemStats)
			admin.POST("/collections/:id/rebuild-index", s.rebuildCollectionIndex)
		}
	}
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		s.logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", c.GetString("request_id")),
		)
	}
}

func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *Server) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

func (s *Server) maxRequestSizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.config.MaxRequestSize)
		c.Next()
	}
}

func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If rate limiter is not configured, allow all requests
		if s.rateLimiter == nil {
			c.Next()
			return
		}

		// Use the rate limiter's middleware with user-based key
		limiterMiddleware := s.rateLimiter.Middleware(ratelimit.UserBasedKey)
		limiterMiddleware(c)
	}
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &security.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.config.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*security.TokenClaims); ok && token.Valid {
			// Set claims in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
	}
}

func (s *Server) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found in context"})
			c.Abort()
			return
		}

		userRole, ok := role.(models.UserRole)
		if !ok {
			// Try string conversion if somehow it lost type info (unlikely with shared models)
			if roleStr, ok := role.(string); ok {
				userRole = models.UserRole(roleStr)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role format"})
				c.Abort()
				return
			}
		}

		if userRole != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Handler placeholders
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (s *Server) readinessCheck(c *gin.Context) {
	// Note: For API Gateway, we check backend service connectivity
	// In a production setup, you'd initialize health.Checker with actual clients
	// For now, return a simple ready status
	// TODO: Add actual dependency checks when centralizing database/redis clients
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "api-gateway",
	})
}

// Service URLs (Hardcoded for local dev or env vars)
var (
	UserServiceUrl       = getEnv("USER_SERVICE_URL", "http://localhost:8086")
	DocumentServiceUrl   = getEnv("DOCUMENT_SERVICE_URL", "http://localhost:8081")
	CollectionServiceUrl = getEnv("COLLECTION_SERVICE_URL", "http://localhost:8082")
	SearchServiceUrl     = getEnv("SEARCH_SERVICE_URL", "http://localhost:8084")
	AnalyticsServiceUrl  = getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8087")
)

// Helper for reverse proxy
func (s *Server) proxyRequest(c *gin.Context, target string, pathRewrite func(string) string) {
	remote, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host

		// Apply path rewrite if provided, otherwise keep as is
		if pathRewrite != nil {
			req.URL.Path = pathRewrite(c.Request.URL.Path)
		} else {
			req.URL.Path = c.Request.URL.Path
		}
		req.URL.RawQuery = c.Request.URL.RawQuery
	}

	// Strip Access-Control headers from backend response to avoid duplicates
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Credentials")
		return nil
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// Rewrites
// Rewrites
func rewriteIdentity(path string) string { return path }

func rewriteDocuments(path string) string {
	return path
}
func rewriteCollections(path string) string {
	return path
}
func rewriteSearch(path string) string {
	// Assume Search Service uses compatible /api/v1/search path
	return path
}

// User rewrites
func rewriteRegister(path string) string { return "/api/v1/auth/register" }
func rewriteLogin(path string) string    { return "/api/v1/auth/login" }
func rewriteProfile(path string) string  { return "/api/v1/auth/me" }

// Admin
func rewriteStats(path string) string { return "/api/v1/stats" }

// Handlers
func (s *Server) uploadDocument(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}
func (s *Server) getDocument(c *gin.Context) { s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments) }
func (s *Server) updateDocument(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}
func (s *Server) deleteDocument(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}
func (s *Server) listDocuments(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}
func (s *Server) downloadDocument(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}
func (s *Server) viewDocument(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}
func (s *Server) getThumbnail(c *gin.Context) {
	s.proxyRequest(c, DocumentServiceUrl, rewriteDocuments)
}

func (s *Server) search(c *gin.Context)         { s.proxyRequest(c, SearchServiceUrl, rewriteSearch) }
func (s *Server) advancedSearch(c *gin.Context) { s.proxyRequest(c, SearchServiceUrl, rewriteSearch) }

func (s *Server) createCollection(c *gin.Context) {
	s.proxyRequest(c, CollectionServiceUrl, rewriteCollections)
}
func (s *Server) getCollection(c *gin.Context) {
	s.proxyRequest(c, CollectionServiceUrl, rewriteCollections)
}
func (s *Server) updateCollection(c *gin.Context) {
	s.proxyRequest(c, CollectionServiceUrl, rewriteCollections)
}
func (s *Server) deleteCollection(c *gin.Context) {
	s.proxyRequest(c, CollectionServiceUrl, rewriteCollections)
}
func (s *Server) listCollections(c *gin.Context) {
	s.proxyRequest(c, CollectionServiceUrl, rewriteCollections)
}
func (s *Server) getCollectionDocuments(c *gin.Context) {
	s.proxyRequest(c, CollectionServiceUrl, rewriteCollections)
}

func (s *Server) registerUser(c *gin.Context)   { s.proxyRequest(c, UserServiceUrl, rewriteRegister) }
func (s *Server) loginUser(c *gin.Context)      { s.proxyRequest(c, UserServiceUrl, rewriteLogin) }
func (s *Server) getUserProfile(c *gin.Context) { s.proxyRequest(c, UserServiceUrl, rewriteProfile) }
func (s *Server) updateUserProfile(c *gin.Context) {
	// Update profile usually PUT /users/:id. We need ID from token.
	// But authHandler has /auth/me or ???
	// Using userHandler PUT /users/:id
	userID := c.GetString("user_id")
	s.proxyRequest(c, UserServiceUrl, func(p string) string { return "/users/" + userID })
}

func (s *Server) deleteUser(c *gin.Context) {
	// Identity rewrite for /users/:id
	s.proxyRequest(c, UserServiceUrl, func(path string) string {
		return path
	})
}

func (s *Server) getSystemStats(c *gin.Context) { s.proxyRequest(c, AnalyticsServiceUrl, rewriteStats) }

// Rebuild index triggers indexer or search service?
// Usually indexer doesn't expose HTTP. Search Service might have admin endpoint?
// Or we trigger via Kafka?
// For now, return 501
func (s *Server) rebuildCollectionIndex(c *gin.Context) { c.JSON(http.StatusNotImplemented, nil) }

func (s *Server) Start() error {
	s.logger.Info("starting server", zap.String("port", s.config.Port))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.server.Shutdown(ctx)
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	config := &Config{
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ShutdownTimeout:   15 * time.Second,
		MaxRequestSize:    100 * 1024 * 1024, // 100MB
		RateLimitRequests: 1000,
		RateLimitWindow:   time.Minute,
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
	}

	server := NewServer(config, logger)

	// Graceful shutdown
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
