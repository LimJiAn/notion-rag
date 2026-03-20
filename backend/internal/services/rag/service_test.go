package rag

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
)

func TestBuildNoContextAnswer(t *testing.T) {
	answer := buildNoContextAnswer("뭐 했지?")

	if answer.UsedContext {
		t.Fatalf("expected no context answer to mark UsedContext=false")
	}
	if answer.ConfidenceLabel != "low" {
		t.Fatalf("expected low confidence, got %q", answer.ConfidenceLabel)
	}
	if len(answer.FollowUpQuestions) != 3 {
		t.Fatalf("expected 3 follow-up questions, got %d", len(answer.FollowUpQuestions))
	}
}

func TestBuildAnswerStripsVectorsAndAddsCitations(t *testing.T) {
	results := []knowledge.SearchResult{
		{
			Document: knowledge.Document{
				Title:     "주간 회의",
				PageID:    "2d105bf2-ff64-80c6-9fd9-eb2f39a07bc8",
				Content:   "다음 주까지 API 정리와 테스트 추가를 진행하기로 했다.",
				Chunk:     1,
				Vector:    []float64{0.1, 0.2, 0.3},
				UpdatedAt: time.Date(2026, 3, 20, 3, 0, 0, 0, time.UTC),
			},
			Similarity: 0.91,
		},
	}

	answer := buildAnswer("무슨 일을 하기로 했지?", "테스트 추가와 API 정리를 진행하기로 했습니다.", results)

	if !answer.UsedContext {
		t.Fatalf("expected grounded answer to mark UsedContext=true")
	}
	if len(answer.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(answer.Results))
	}
	if len(answer.Citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(answer.Citations))
	}
	if answer.Results[0].Document.NotionURL == "" {
		t.Fatalf("expected notion url to be populated")
	}
	if answer.ConfidenceLabel != "high" {
		t.Fatalf("expected high confidence, got %q", answer.ConfidenceLabel)
	}

	body, err := json.Marshal(answer)
	if err != nil {
		t.Fatalf("marshal answer: %v", err)
	}
	if strings.Contains(string(body), "\"vector\"") {
		t.Fatalf("query response should not expose embedding vectors: %s", string(body))
	}
}
