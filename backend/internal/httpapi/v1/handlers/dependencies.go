package handlers

import (
	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
	ingestservice "github.com/jian1990/notion-rag/backend/internal/services/ingest"
	ragservice "github.com/jian1990/notion-rag/backend/internal/services/rag"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type Dependencies struct {
	Store    knowledge.Store
	Ingest   *ingestservice.Service
	RAG      *ragservice.Service
	Settings *settings.Store
}
