package httpapi

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func (s *Server) Engine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	engine.GET("/healthz", s.handleHealth)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/stats", s.handleStats)
		v1.POST("/sync", s.handleSync)
		v1.POST("/query", s.handleQuery)
		v1.GET("/settings", s.handleGetSettings)
		v1.PUT("/settings", s.handleUpdateSettings)
	}

	return engine
}
