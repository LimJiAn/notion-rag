import type { FormEvent, ReactNode } from "react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Textarea } from "../../components/ui/textarea";
import { formatConfidence, formatDate, openNotionPage, resolveNotionURL, trimSnippet } from "../../lib/format";
import type { ChatMessage, SearchResult } from "../../lib/types";
import { cn } from "../../lib/utils";

export function AssistantLayout({
  history,
  question,
  queryStatus,
  asking,
  selectedSource,
  visibleResults,
  lastAssistantMessage,
  onQuestionChange,
  onQuestionSubmit,
  onSourceSelect,
  onSuggestedQuestionSelect,
}: {
  history: ChatMessage[];
  question: string;
  queryStatus: string;
  asking: boolean;
  selectedSource: SearchResult | null;
  visibleResults: SearchResult[];
  lastAssistantMessage: ChatMessage | null;
  onQuestionChange: (value: string) => void;
  onQuestionSubmit: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  onSourceSelect: (value: SearchResult) => void;
  onSuggestedQuestionSelect: (value: string) => void;
}) {
  const selectedSourceURL = selectedSource
    ? resolveNotionURL(selectedSource.document.page_id, selectedSource.document.notion_url)
    : null;

  return (
    <main className="grid flex-1 gap-5 xl:grid-cols-[minmax(0,1.08fr)_minmax(340px,0.92fr)]">
      <Card className="flex min-h-[720px] flex-col p-6">
        <CardHeader className="pb-5">
          <CardTitle>대화</CardTitle>
          <CardDescription>노션에서 찾은 문맥을 바탕으로 답변합니다.</CardDescription>
        </CardHeader>

        <CardContent className="flex flex-1 flex-col gap-5 p-0">
          <div className="flex-1 space-y-3 overflow-auto pr-1">
            {history.length === 0 ? (
              <EmptyState>아직 대화가 없습니다. 먼저 Notion을 동기화한 뒤 질문해 보세요.</EmptyState>
            ) : (
              history.map((message) => <MessageItem key={message.id} message={message} />)
            )}
          </div>

          <form className="space-y-4 border-t border-ink-100 pt-5" onSubmit={onQuestionSubmit}>
            <Textarea
              id="question"
              name="question"
              rows={6}
              value={question}
              onChange={(event) => onQuestionChange(event.target.value)}
              placeholder="예: 지난주 회의에서 정리한 우선순위를 요약해줘"
            />
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-ink-500">{queryStatus}</p>
              <Button size="lg" disabled={asking} type="submit">
                {asking ? "생성 중..." : "질문 보내기"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <section className="grid gap-5">
        <Card className="p-6">
          <CardHeader className="pb-4">
            <CardTitle>최근 답변</CardTitle>
            <CardDescription>가장 최근에 생성된 응답을 보여줍니다.</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            {!lastAssistantMessage ? (
              <EmptyState>최근 답변이 여기에 표시됩니다.</EmptyState>
            ) : (
              <div className="space-y-3">
                <p className="font-display text-sm font-semibold uppercase tracking-[0.2em] text-ember-500">
                  {findRelatedQuestion(history, lastAssistantMessage.id) ?? "최근 답변"}
                </p>
                <div className="flex flex-wrap gap-2">
                  <Badge className={confidenceBadgeClass(lastAssistantMessage.confidenceLabel)}>
                    {formatConfidence(lastAssistantMessage.confidenceLabel, lastAssistantMessage.confidenceScore)}
                  </Badge>
                  <Badge className={lastAssistantMessage.usedContext ? "bg-moss-100 text-moss-700" : "bg-ember-100 text-ember-700"}>
                    {lastAssistantMessage.usedContext ? "근거 기반 답변" : "근거 부족"}
                  </Badge>
                </div>
                <div className="whitespace-pre-wrap text-sm leading-7 text-ink-900">
                  {lastAssistantMessage.text}
                </div>
                {lastAssistantMessage.citations && lastAssistantMessage.citations.length > 0 ? (
                  <div className="space-y-2 rounded-[24px] border border-ink-100 bg-white/75 p-4">
                    <p className="font-display text-xs font-semibold uppercase tracking-[0.18em] text-ink-500">
                      핵심 근거
                    </p>
                    <div className="space-y-3">
                      {lastAssistantMessage.citations.map((citation) => (
                        <button
                          className="block w-full rounded-2xl border border-ink-100 bg-white px-3 py-3 text-left transition hover:border-ember-200 hover:bg-ember-50/40"
                          key={`${citation.page_id}-${citation.similarity}`}
                          onClick={() => openNotionPage(citation.page_id, citation.notion_url)}
                          type="button"
                        >
                          <div className="mb-1 flex items-center justify-between gap-3">
                            <span className="text-sm font-medium text-ink-900">{citation.title || "Untitled"}</span>
                            <Badge>similarity {citation.similarity.toFixed(3)}</Badge>
                          </div>
                          <p className="text-sm leading-6 text-ink-600">{citation.snippet}</p>
                        </button>
                      ))}
                    </div>
                  </div>
                ) : null}
                {lastAssistantMessage.followUpQuestions && lastAssistantMessage.followUpQuestions.length > 0 ? (
                  <div className="space-y-2">
                    <p className="font-display text-xs font-semibold uppercase tracking-[0.18em] text-ink-500">
                      바로 이어서 질문하기
                    </p>
                    <div className="flex flex-wrap gap-2">
                      {lastAssistantMessage.followUpQuestions.map((suggestion) => (
                        <Button
                          className="h-auto rounded-full px-4 py-2 text-left"
                          key={suggestion}
                          onClick={() => onSuggestedQuestionSelect(suggestion)}
                          size="sm"
                          type="button"
                          variant="outline"
                        >
                          {suggestion}
                        </Button>
                      ))}
                    </div>
                  </div>
                ) : null}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="p-6">
          <CardHeader className="pb-4">
            <CardTitle>검색 근거</CardTitle>
            <CardDescription>클릭하면 선택되고, 더블클릭하면 Notion 원문이 열립니다.</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            {visibleResults.length === 0 ? (
              <EmptyState>답변을 생성하면 관련 문서가 여기에 표시됩니다.</EmptyState>
            ) : (
              <div className="space-y-3">
                {visibleResults.map((item) => (
                  <button
                    className={cn(
                      "w-full rounded-[24px] border bg-white/70 p-4 text-left shadow-sm transition hover:-translate-y-0.5 hover:bg-white",
                      selectedSource?.document.page_id === item.document.page_id
                        ? "border-ember-500/40 bg-ember-100/45"
                        : "border-ink-100",
                    )}
                    key={`${item.document.page_id}-${item.document.title}-${item.similarity}`}
                    onClick={() => onSourceSelect(item)}
                    onDoubleClick={() => openNotionPage(item.document.page_id, item.document.notion_url)}
                    type="button"
                  >
                    <div className="mb-2 flex items-start justify-between gap-4">
                      <h3 className="font-medium text-ink-900">{item.document.title || "Untitled"}</h3>
                      <span className="font-display text-xs font-semibold uppercase tracking-[0.18em] text-ember-500">
                        Open
                      </span>
                    </div>
                    <div className="mb-3 flex flex-wrap items-center justify-between gap-2 text-xs text-ink-500">
                      <span>page_id: {item.document.page_id}</span>
                      <Badge>similarity {item.similarity.toFixed(3)}</Badge>
                    </div>
                    {item.document.updated_at ? (
                      <p className="mb-2 text-xs text-ink-500">updated {formatDate(item.document.updated_at)}</p>
                    ) : null}
                    <p className="text-sm leading-6 text-ink-700">{trimSnippet(item.document.content)}</p>
                  </button>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="p-6">
          <CardHeader className="pb-4">
            <CardTitle>선택한 문서</CardTitle>
            <CardDescription>선택한 검색 근거의 본문을 확인합니다.</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            {!selectedSource ? (
              <EmptyState>검색 근거를 선택하면 문서 내용을 자세히 볼 수 있습니다.</EmptyState>
            ) : (
              <div className="space-y-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="space-y-1">
                    <h3 className="text-lg font-semibold text-ink-900">
                      {selectedSource.document.title || "Untitled"}
                    </h3>
                    <p className="text-sm text-ink-500">page_id: {selectedSource.document.page_id}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge>similarity {selectedSource.similarity.toFixed(3)}</Badge>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        openNotionPage(selectedSource.document.page_id, selectedSource.document.notion_url)
                      }
                      type="button"
                    >
                      Notion에서 열기
                    </Button>
                  </div>
                </div>
                {selectedSourceURL ? (
                  <a
                    className="block break-all text-sm text-ember-500 hover:underline"
                    href={selectedSourceURL}
                    rel="noreferrer"
                    target="_blank"
                  >
                    {selectedSourceURL}
                  </a>
                ) : null}
                <div className="rounded-[24px] border border-ink-100 bg-white/75 p-4 text-sm leading-7 text-ink-800">
                  {selectedSource.document.content}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </section>
    </main>
  );
}

function MessageItem({ message }: { message: ChatMessage }) {
  return (
    <article
      className={cn(
        "rounded-[26px] border px-4 py-4 shadow-sm",
        message.role === "user" ? "border-ember-100 bg-ember-100/60" : "border-moss-100 bg-moss-100/55",
      )}
    >
      <div className="mb-2 flex items-center justify-between gap-4 text-xs text-ink-500">
        <span className="font-display font-semibold uppercase tracking-[0.18em]">
          {message.role === "user" ? "You" : "Assistant"}
        </span>
        <time dateTime={message.createdAt}>{formatDate(message.createdAt)}</time>
      </div>
      <p className="whitespace-pre-wrap text-sm leading-7 text-ink-900">{message.text}</p>
    </article>
  );
}

function EmptyState({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-[26px] border border-dashed border-ink-200 bg-white/65 px-5 py-8 text-sm leading-7 text-ink-500">
      {children}
    </div>
  );
}

function findRelatedQuestion(history: ChatMessage[], assistantId: string) {
  const index = history.findIndex((entry) => entry.id === assistantId);
  if (index <= 0) {
    return null;
  }

  for (let current = index - 1; current >= 0; current -= 1) {
    if (history[current].role === "user") {
      return history[current].text;
    }
  }
  return null;
}

function confidenceBadgeClass(label?: "high" | "medium" | "low") {
  switch (label) {
    case "high":
      return "bg-moss-100 text-moss-700";
    case "medium":
      return "bg-amber-100 text-amber-700";
    case "low":
      return "bg-ember-100 text-ember-700";
    default:
      return "";
  }
}
