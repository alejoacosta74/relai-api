package api

import (
	"context"
	"log"
	"net/http"

	"relai/internal/client"
	"relai/internal/config"

	"github.com/gin-gonic/gin"
)

// Server wraps the Gin engine and the underlying HTTP server.
type Server struct {
	Engine *gin.Engine
	srv    *http.Server
	client client.KrakenClientInterface
}

// NewServer creates a new Server instance with middleware and routes setup.
func NewServer(cfg *config.Config) *Server {
	engine := gin.New()

	// Register middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

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
