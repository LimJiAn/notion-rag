package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/jian1990/notion-rag/backend/internal/httpapi/middleware"
	v1handlers "github.com/jian1990/notion-rag/backend/internal/httpapi/v1/handlers"
)

func New(deps v1handlers.Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	healthHandler := v1handlers.NewHealthHandler(deps)
	ragHandler := v1handlers.NewRAGHandler(deps)
	settingsHandler := v1handlers.NewSettingsHandler(deps)
	knowledgeHandler := v1handlers.NewKnowledgeHandler(deps)

	engine.GET("/healthz", healthHandler.GetHealth)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/stats", healthHandler.GetStats)
		v1.POST("/sync", ragHandler.Sync)
		v1.POST("/query", ragHandler.Query)
		v1.GET("/documents", knowledgeHandler.ListDocuments)
		v1.GET("/settings", settingsHandler.Get)
		v1.PUT("/settings", settingsHandler.Update)
	}

	return engine
}
