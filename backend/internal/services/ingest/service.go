package ingest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/chunk"
	"github.com/jian1990/notion-rag/backend/internal/clients/gemini"
	"github.com/jian1990/notion-rag/backend/internal/clients/notion"
	"github.com/jian1990/notion-rag/backend/internal/config"
	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
	"github.com/jian1990/notion-rag/backend/internal/repositories/documents"
	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type Service struct {
	cfg      config.Config
	settings *settings.Store
	notion   *notion.Client
	embed    *gemini.EmbedClient
	store    *documents.Store
}

func NewService(cfg config.Config, settingsStore *settings.Store, notionClient *notion.Client, embedClient *gemini.EmbedClient, store *documents.Store) *Service {
	return &Service{
		cfg:      cfg,
		settings: settingsStore,
		notion:   notionClient,
		embed:    embedClient,
		store:    store,
	}
}

func (s *Service) Sync(ctx context.Context) (map[string]any, error) {
	pages, err := s.notion.Crawl(ctx, s.settings.Snapshot().NotionRootPageIDs)
	if err != nil {
		return nil, err
	}

	documentsToIndex := make([]knowledge.Document, 0)
	for _, page := range pages {
		chunks := chunk.Text(page.Title+"\n"+page.Content, s.cfg.ChunkSize, s.cfg.ChunkOverlap)
		for idx, part := range chunks {
			documentsToIndex = append(documentsToIndex, knowledge.Document{
				ID:      fmt.Sprintf("%s-%d", page.ID, idx),
				PageID:  page.ID,
				Title:   page.Title,
				Content: part,
				Chunk:   idx,
				Metadata: map[string]string{
					"source": "notion",
				},
				UpdatedAt: time.Now().UTC(),
			})
		}
	}

	indexed, err := s.embedDocuments(ctx, documentsToIndex)
	if err != nil {
		return nil, err
	}
	if err := s.store.Replace(ctx, indexed); err != nil {
		return nil, err
	}

	stats := s.store.Stats()
	stats["pages"] = len(pages)
	stats["chunks"] = len(indexed)
	return stats, nil
}

func (s *Service) embedDocuments(ctx context.Context, docs []knowledge.Document) ([]knowledge.Document, error) {
	type job struct {
		index int
		doc   knowledge.Document
	}
	type result struct {
		index int
		doc   knowledge.Document
		err   error
	}

	jobs := make(chan job)
	results := make(chan result, len(docs))
	var wg sync.WaitGroup

	for i := 0; i < s.cfg.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				vector, err := s.embed.Embed(ctx, item.doc.Content)
				item.doc.Vector = vector
				results <- result{index: item.index, doc: item.doc, err: err}
			}
		}()
	}

	go func() {
		for index, doc := range docs {
			select {
			case <-ctx.Done():
				break
			case jobs <- job{index: index, doc: doc}:
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	indexed := make([]knowledge.Document, len(docs))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		indexed[res.index] = res.doc
	}
	return indexed, nil
}
