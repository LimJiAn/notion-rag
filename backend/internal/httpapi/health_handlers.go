package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleHealth returns the server health status.
//
// @Summary Health check
// @Description Returns the backend health state
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleStats returns current index statistics.
//
// @Summary Index stats
// @Description Returns indexed document and last sync stats
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/stats [get]
func (s *Server) handleStats(c *gin.Context) {
	c.JSON(http.StatusOK, s.store.Stats())
}
