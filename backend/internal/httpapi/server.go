package httpapi

import (
	"github.com/jian1990/notion-rag/backend/internal/repositories/documents"
	ingestservice "github.com/jian1990/notion-rag/backend/internal/services/ingest"
	ragservice "github.com/jian1990/notion-rag/backend/internal/services/rag"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type Server struct {
	store    *documents.Store
	ingest   *ingestservice.Service
	rag      *ragservice.Service
	settings *settings.Store
}

func NewServer(store *documents.Store, ingest *ingestservice.Service, rag *ragservice.Service, settingsStore *settings.Store) *Server {
	return &Server{
		store:    store,
		ingest:   ingest,
		rag:      rag,
		settings: settingsStore,
	}
}
