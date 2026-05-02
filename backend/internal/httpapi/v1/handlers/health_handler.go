package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
)

type HealthHandler struct {
	store knowledge.Store
}

func NewHealthHandler(deps Dependencies) *HealthHandler {
	return &HealthHandler{store: deps.Store}
}

// GetHealth returns the server health status.
//
// @Summary Health check
// @Description Returns the backend health state
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetStats returns current index statistics.
//
// @Summary Index stats
// @Description Returns indexed document and last sync stats
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/stats [get]
func (h *HealthHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.Stats())
}
