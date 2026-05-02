package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/httpapi/v1/responses"
)

type KnowledgeHandler struct {
	deps Dependencies
}

func NewKnowledgeHandler(deps Dependencies) *KnowledgeHandler {
	return &KnowledgeHandler{deps: deps}
}

// ListDocuments returns indexed knowledge chunks.
//
// @Summary List indexed documents
// @Description Returns recent indexed Notion chunks without exposing embedding vectors
// @Tags knowledge
// @Produce json
// @Param limit query int false "Maximum number of documents to return"
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} responses.ErrorResponse
// @Router /api/v1/documents [get]
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	limit := parseLimit(c.Query("limit"), 25)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	documents, err := h.deps.Store.List(ctx, limit)
	if err != nil {
		responses.WriteError(c, http.StatusBadGateway, err.Error())
		return
	}

	out := make([]responses.DocumentResponse, 0, len(documents))
	for _, doc := range documents {
		out = append(out, responses.NewDocumentResponse(doc))
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(out),
		"documents": out,
	})
}

func parseLimit(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	if parsed > 100 {
		return 100
	}
	return parsed
}
