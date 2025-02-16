package api

import (
	"context"
	"log"
	"net/http"

	"relai/internal/api/middleware"
	"relai/internal/client"
	"relai/internal/config"

	"github.com/alejoacosta74/go-logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Server wraps the Gin engine and the underlying HTTP server.
type Server struct {
	Engine *gin.Engine
	srv    *http.Server
	client client.KrakenClientInterface
}

// NewServer creates a new Server instance with middleware and routes setup.
func NewServer(cfg *config.Config) *Server {
	if cfg.LogLevel == "debug" || cfg.LogLevel == "trace" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Register middleware
	engine.Use(gin.Recovery())
	if cfg.LogLevel == "debug" || cfg.LogLevel == "trace" {
		engine.Use(gin.Logger())
	}

	rateLimiter, err := middleware.NewIPRateLimiter(
		rate.Limit(cfg.RateLimit.RequestsPerMinute/60),
		cfg.RateLimit.BurstSize,
		cfg.RateLimit.CacheSize, // e.g., 10000
	)
	if err != nil {
		logger.Fatalf("Failed to create rate limiter with given config (%+v): %v", cfg.RateLimit, err)
	}

	engine.Use(middleware.RateLimit(rateLimiter))

	krakenClient := client.NewKrakenClient(cfg.KrakenAPIBaseEndpoint)

	// Create the underlying HTTP server.
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	server := &Server{
		Engine: engine,
		srv:    srv,
		client: krakenClient,
	}

	server.registerRoutes()

	return server
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	log.Printf("Server is starting on %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server with the provided context.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server gracefully...")
	return s.srv.Shutdown(ctx)
}
