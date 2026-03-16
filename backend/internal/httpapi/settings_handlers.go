package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

// handleGetSettings returns current runtime settings metadata.
//
// @Summary Get runtime settings
// @Description Returns stored runtime settings metadata without exposing secret values
// @Tags settings
// @Produce json
// @Success 200 {object} settings.PublicValues
// @Router /api/v1/settings [get]
func (s *Server) handleGetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, s.settings.Public())
}

// handleUpdateSettings updates runtime settings.
//
// @Summary Update runtime settings
// @Description Stores runtime settings in the backend data volume
// @Tags settings
// @Accept json
// @Produce json
// @Param payload body updateSettingsRequest true "Settings payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/settings [put]
func (s *Server) handleUpdateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid json body")
		return
	}

	err := s.settings.Update(settings.UpdateInput{
		NotionToken:       req.NotionToken,
		NotionVersion:     req.NotionVersion,
		NotionRootPageIDs: splitCSV(req.NotionRootPageIDs),
		GeminiAPIKey:      req.GeminiAPIKey,
		EmbeddingModel:    req.EmbeddingModel,
		GenerationModel:   req.GenerationModel,
	})
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "saved",
		"settings": s.settings.Public(),
	})
}
