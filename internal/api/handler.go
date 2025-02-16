package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerRoutes sets up the API routes
func (s *Server) registerRoutes() {
	s.Engine.GET("/api/v1/ltp", s.handleGetLTP)
}

// handleGetLTP handles the LTP endpoint
func (s *Server) handleGetLTP(c *gin.Context) {
	response, err := s.client.GetLTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
