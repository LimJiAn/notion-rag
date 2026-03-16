package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/httpapi/v1/requests"
	"github.com/jian1990/notion-rag/backend/internal/httpapi/v1/responses"
	ingestservice "github.com/jian1990/notion-rag/backend/internal/services/ingest"
	ragservice "github.com/jian1990/notion-rag/backend/internal/services/rag"
)

type RAGHandler struct {
	ingest *ingestservice.Service
	rag    *ragservice.Service
}

func NewRAGHandler(deps Dependencies) *RAGHandler {
	return &RAGHandler{
		ingest: deps.Ingest,
		rag:    deps.RAG,
	}
}

// Sync syncs Notion pages into the local index.
//
// @Summary Sync Notion content
// @Description Crawls Notion pages, chunks them, embeds them, and stores them locally
// @Tags rag
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} responses.ErrorResponse
// @Router /api/v1/sync [post]
func (h *RAGHandler) Sync(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	stats, err := h.ingest.Sync(ctx)
	if err != nil {
		responses.WriteError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "synced",
		"stats":  stats,
	})
}

// Query answers a question using indexed Notion content.
//
// @Summary Query indexed knowledge
// @Description Searches indexed Notion chunks and generates an answer
// @Tags rag
// @Accept json
// @Produce json
// @Param payload body requests.QueryRequest true "Question payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 502 {object} responses.ErrorResponse
// @Router /api/v1/query [post]
func (h *RAGHandler) Query(c *gin.Context) {
	var req requests.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, http.StatusBadRequest, "invalid json body")
		return
	}
	if strings.TrimSpace(req.Question) == "" {
		responses.WriteError(c, http.StatusBadRequest, "question is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	answer, err := h.rag.Ask(ctx, req.Question)
	if err != nil {
		responses.WriteError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, answer)
}
