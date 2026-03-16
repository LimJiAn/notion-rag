package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/jian1990/notion-rag/backend/internal/clients/gemini"
	"github.com/jian1990/notion-rag/backend/internal/config"
	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
	"github.com/jian1990/notion-rag/backend/internal/repositories/documents"
)

type Service struct {
	cfg      config.Config
	store    *documents.Store
	embed    *gemini.EmbedClient
	generate *gemini.GenerateClient
}

type Answer struct {
	Question string                   `json:"question"`
	Answer   string                   `json:"answer"`
	Results  []knowledge.SearchResult `json:"results"`
}

func NewService(cfg config.Config, store *documents.Store, embedClient *gemini.EmbedClient, generateClient *gemini.GenerateClient) *Service {
	return &Service{
		cfg:      cfg,
		store:    store,
		embed:    embedClient,
		generate: generateClient,
	}
}

func (s *Service) Ask(ctx context.Context, question string) (Answer, error) {
	vector, err := s.embed.Embed(ctx, question)
	if err != nil {
		return Answer{}, err
	}

	results, err := s.store.Search(ctx, vector, s.cfg.TopK, s.cfg.SimilarityCutoff)
	if err != nil {
		return Answer{}, err
	}

	answer, err := s.generate.Answer(ctx, buildPrompt(question, results))
	if err != nil {
		return Answer{}, err
	}

	return Answer{
		Question: question,
		Answer:   answer,
		Results:  results,
	}, nil
}

func buildPrompt(question string, results []knowledge.SearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("질문: %s\n\n관련 문맥이 없습니다. 지어내지 말고 정보가 부족하다고 답하세요.", question)
	}

	var contextParts []string
	for _, result := range results {
		contextParts = append(contextParts, fmt.Sprintf(
			"[문서: %s | page_id=%s | similarity=%.3f]\n%s",
			result.Document.Title,
			result.Document.PageID,
			result.Similarity,
			result.Document.Content,
		))
	}

	return fmt.Sprintf(`당신은 사용자의 Notion 개인 비서입니다.
아래 문맥만 기반으로 답하세요.
문맥에 없는 내용은 지어내지 말고 "정보가 부족하여 알 수 없습니다"라고 답하세요.
가능하면 어떤 문서에서 근거를 찾았는지 함께 설명하세요.

[Context]
%s

[Question]
%s`, strings.Join(contextParts, "\n\n"), question)
}
