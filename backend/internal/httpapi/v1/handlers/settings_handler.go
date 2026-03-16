package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/httpapi/v1/requests"
	"github.com/jian1990/notion-rag/backend/internal/httpapi/v1/responses"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type SettingsHandler struct {
	settings *settings.Store
}

func NewSettingsHandler(deps Dependencies) *SettingsHandler {
	return &SettingsHandler{settings: deps.Settings}
}

// Get returns current runtime settings metadata.
//
// @Summary Get runtime settings
// @Description Returns stored runtime settings metadata without exposing secret values
// @Tags settings
// @Produce json
// @Success 200 {object} settings.PublicValues
// @Router /api/v1/settings [get]
func (h *SettingsHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, h.settings.Public())
}

// Update updates runtime settings.
//
// @Summary Update runtime settings
// @Description Stores runtime settings in the backend data volume
// @Tags settings
// @Accept json
// @Produce json
// @Param payload body requests.UpdateSettingsRequest true "Settings payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Router /api/v1/settings [put]
func (h *SettingsHandler) Update(c *gin.Context) {
	var req requests.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, http.StatusBadRequest, "invalid json body")
		return
	}

	err := h.settings.Update(settings.UpdateInput{
		NotionToken:       req.NotionToken,
		NotionVersion:     req.NotionVersion,
		NotionRootPageIDs: req.RootPageIDs(),
		GeminiAPIKey:      req.GeminiAPIKey,
		EmbeddingModel:    req.EmbeddingModel,
		GenerationModel:   req.GenerationModel,
	})
	if err != nil {
		responses.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "saved",
		"settings": h.settings.Public(),
	})
}
