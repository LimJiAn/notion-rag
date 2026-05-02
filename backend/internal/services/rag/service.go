package rag

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/clients/gemini"
	"github.com/jian1990/notion-rag/backend/internal/config"
	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
)

type Service struct {
	cfg      config.Config
	store    knowledge.Store
	embed    *gemini.EmbedClient
	generate *gemini.GenerateClient
}

type SourceDocument struct {
	Title     string `json:"title"`
	PageID    string `json:"page_id"`
	Content   string `json:"content"`
	Chunk     int    `json:"chunk"`
	NotionURL string `json:"notion_url,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type SearchResult struct {
	Similarity float64        `json:"similarity"`
	Document   SourceDocument `json:"document"`
}

type Citation struct {
	Title      string  `json:"title"`
	PageID     string  `json:"page_id"`
	NotionURL  string  `json:"notion_url,omitempty"`
	Snippet    string  `json:"snippet"`
	Similarity float64 `json:"similarity"`
}

type Answer struct {
	Question          string         `json:"question"`
	Answer            string         `json:"answer"`
	Results           []SearchResult `json:"results"`
	Citations         []Citation     `json:"citations"`
	ConfidenceScore   float64        `json:"confidence_score"`
	ConfidenceLabel   string         `json:"confidence_label"`
	UsedContext       bool           `json:"used_context"`
	FollowUpQuestions []string       `json:"follow_up_questions"`
}

func NewService(cfg config.Config, store knowledge.Store, embedClient *gemini.EmbedClient, generateClient *gemini.GenerateClient) *Service {
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

	if len(results) == 0 {
		return buildNoContextAnswer(question), nil
	}

	answer, err := s.generate.Answer(ctx, buildPrompt(question, results))
	if err != nil {
		return Answer{}, err
	}

	return buildAnswer(question, answer, results), nil
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

func buildAnswer(question, answer string, results []knowledge.SearchResult) Answer {
	confidenceScore, confidenceLabel := deriveConfidence(results)

	return Answer{
		Question:          question,
		Answer:            answer,
		Results:           toResponseResults(results),
		Citations:         buildCitations(results),
		ConfidenceScore:   confidenceScore,
		ConfidenceLabel:   confidenceLabel,
		UsedContext:       true,
		FollowUpQuestions: buildFollowUpQuestions(results),
	}
}

func buildNoContextAnswer(question string) Answer {
	return Answer{
		Question:        question,
		Answer:          "현재 동기화된 Notion 문서에서 직접적인 근거를 찾지 못했습니다. 기간, 프로젝트명, 문서명 같은 키워드를 더 넣어서 다시 질문해 주세요.",
		ConfidenceScore: 0.18,
		ConfidenceLabel: "low",
		UsedContext:     false,
		FollowUpQuestions: []string{
			"지난주나 특정 기간을 넣어서 다시 질문해줘",
			"프로젝트명이나 문서 제목을 포함해서 다시 질문해줘",
			"회의 메모, 업무 기록, 아이디어 중 어떤 내용인지 지정해줘",
		},
	}
}

func toResponseResults(results []knowledge.SearchResult) []SearchResult {
	out := make([]SearchResult, 0, len(results))
	for _, result := range results {
		out = append(out, SearchResult{
			Similarity: result.Similarity,
			Document: SourceDocument{
				Title:     result.Document.Title,
				PageID:    result.Document.PageID,
				Content:   result.Document.Content,
				Chunk:     result.Document.Chunk,
				NotionURL: buildNotionURL(result.Document.PageID),
				UpdatedAt: formatUpdatedAt(result.Document.UpdatedAt),
			},
		})
	}
	return out
}

func buildCitations(results []knowledge.SearchResult) []Citation {
	limit := min(3, len(results))
	out := make([]Citation, 0, limit)
	for i := 0; i < limit; i++ {
		doc := results[i].Document
		out = append(out, Citation{
			Title:      doc.Title,
			PageID:     doc.PageID,
			NotionURL:  buildNotionURL(doc.PageID),
			Snippet:    trimSnippet(doc.Content, 180),
			Similarity: results[i].Similarity,
		})
	}
	return out
}

func deriveConfidence(results []knowledge.SearchResult) (float64, string) {
	if len(results) == 0 {
		return 0.18, "low"
	}

	limit := min(3, len(results))
	var weightedTotal float64
	var weightSum float64

	for i := 0; i < limit; i++ {
		weight := 1.0
		if i == 0 {
			weight = 1.4
		}
		weightedTotal += results[i].Similarity * weight
		weightSum += weight
	}

	score := math.Min(1, weightedTotal/weightSum)

	switch {
	case score >= 0.85:
		return score, "high"
	case score >= 0.74:
		return score, "medium"
	default:
		return score, "low"
	}
}

func buildFollowUpQuestions(results []knowledge.SearchResult) []string {
	if len(results) == 0 {
		return nil
	}

	suggestions := make([]string, 0, 3)
	if title := strings.TrimSpace(results[0].Document.Title); title != "" {
		suggestions = append(suggestions, fmt.Sprintf("%s 문서 기준으로 핵심만 다시 요약해줘", title))
	}
	suggestions = append(suggestions,
		"관련된 액션 아이템만 추려줘",
		"이 내용을 시간순으로 다시 정리해줘",
	)

	return suggestions[:min(3, len(suggestions))]
}

func buildNotionURL(pageID string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(pageID), "-", "")
	if normalized == "" {
		return ""
	}
	return fmt.Sprintf("https://www.notion.so/%s", normalized)
}

func formatUpdatedAt(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func trimSnippet(value string, max int) string {
	clean := strings.Join(strings.Fields(value), " ")
	if len(clean) <= max {
		return clean
	}
	return clean[:max] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
