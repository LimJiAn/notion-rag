package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// handleSync syncs Notion pages into the local index.
//
// @Summary Sync Notion content
// @Description Crawls Notion pages, chunks them, embeds them, and stores them locally
// @Tags rag
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} map[string]string
// @Router /api/v1/sync [post]
func (s *Server) handleSync(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	stats, err := s.ingest.Sync(ctx)
	if err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "synced",
		"stats":  stats,
	})
}

// handleQuery answers a question using indexed Notion content.
//
// @Summary Query indexed knowledge
// @Description Searches indexed Notion chunks and generates an answer
// @Tags rag
// @Accept json
// @Produce json
// @Param payload body queryRequest true "Question payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /api/v1/query [post]
func (s *Server) handleQuery(c *gin.Context) {
	var req queryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json body")
		return
	}
	if strings.TrimSpace(req.Question) == "" {
		writeError(c, http.StatusBadRequest, "question is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	answer, err := s.rag.Ask(ctx, req.Question)
	if err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, answer)
}
