package app

import (
	"net/http"

	"github.com/jian1990/notion-rag/backend/internal/config"
	"github.com/jian1990/notion-rag/backend/internal/embed"
	"github.com/jian1990/notion-rag/backend/internal/generate"
	"github.com/jian1990/notion-rag/backend/internal/httpapi"
	"github.com/jian1990/notion-rag/backend/internal/ingest"
	"github.com/jian1990/notion-rag/backend/internal/notion"
	"github.com/jian1990/notion-rag/backend/internal/rag"
	"github.com/jian1990/notion-rag/backend/internal/settings"
	"github.com/jian1990/notion-rag/backend/internal/store"
)

func NewServer(cfg config.Config) (*http.Server, error) {
	docStore, err := store.New(cfg.StorePath)
	if err != nil {
		return nil, err
	}

	settingsStore, err := settings.New(cfg)
	if err != nil {
		return nil, err
	}

	notionClient := notion.NewClient(settingsStore, cfg.RequestTimeout)
	embedClient := embed.NewClient(settingsStore, cfg.RequestTimeout)
	generateClient := generate.NewClient(settingsStore, cfg.RequestTimeout)

	ingestService := ingest.NewService(cfg, settingsStore, notionClient, embedClient, docStore)
	ragService := rag.NewService(cfg, docStore, embedClient, generateClient)
	apiServer := httpapi.NewServer(docStore, ingestService, ragService, settingsStore)

	return &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: apiServer.Handler(),
	}, nil
}
