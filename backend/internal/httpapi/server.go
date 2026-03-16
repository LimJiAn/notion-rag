package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/httpapi/router"
	v1handlers "github.com/jian1990/notion-rag/backend/internal/httpapi/v1/handlers"
	"github.com/jian1990/notion-rag/backend/internal/repositories/documents"
	ingestservice "github.com/jian1990/notion-rag/backend/internal/services/ingest"
	ragservice "github.com/jian1990/notion-rag/backend/internal/services/rag"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type Server struct {
	engine *gin.Engine
}

func NewServer(store *documents.Store, ingest *ingestservice.Service, rag *ragservice.Service, settingsStore *settings.Store) *Server {
	deps := v1handlers.Dependencies{
		Store:    store,
		Ingest:   ingest,
		RAG:      rag,
		Settings: settingsStore,
	}

	return &Server{
		engine: router.New(deps),
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}
