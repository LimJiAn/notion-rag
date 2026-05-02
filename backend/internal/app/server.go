package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jian1990/notion-rag/backend/internal/clients/gemini"
	"github.com/jian1990/notion-rag/backend/internal/clients/notion"
	"github.com/jian1990/notion-rag/backend/internal/config"
	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
	"github.com/jian1990/notion-rag/backend/internal/httpapi"
	"github.com/jian1990/notion-rag/backend/internal/repositories/documents"
	weaviatestore "github.com/jian1990/notion-rag/backend/internal/repositories/weaviate"
	ingestservice "github.com/jian1990/notion-rag/backend/internal/services/ingest"
	ragservice "github.com/jian1990/notion-rag/backend/internal/services/rag"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

func NewServer(cfg config.Config) (*http.Server, error) {
	docStore, err := newDocumentStore(cfg)
	if err != nil {
		return nil, err
	}

	settingsStore, err := settings.New(cfg)
	if err != nil {
		return nil, err
	}

	notionClient := notion.NewClient(settingsStore, cfg.RequestTimeout)
	embedClient := gemini.NewEmbedClient(settingsStore, cfg.RequestTimeout)
	generateClient := gemini.NewGenerateClient(settingsStore, cfg.RequestTimeout)

	gin.SetMode(gin.ReleaseMode)

	ingestService := ingestservice.NewService(cfg, settingsStore, notionClient, embedClient, docStore)
	ragService := ragservice.NewService(cfg, docStore, embedClient, generateClient)
	apiServer := httpapi.NewServer(docStore, ingestService, ragService, settingsStore)

	return &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: apiServer.Engine(),
	}, nil
}

func newDocumentStore(cfg config.Config) (knowledge.Store, error) {
	if cfg.VectorStore == "weaviate" {
		return weaviatestore.New(cfg.WeaviateURL, cfg.WeaviateClassName, cfg.RequestTimeout)
	}
	return documents.New(cfg.StorePath)
}
