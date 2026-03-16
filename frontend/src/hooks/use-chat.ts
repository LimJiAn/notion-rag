import type { FormEvent } from "react";
import { useEffect, useMemo, useState } from "react";

import { queryKnowledge } from "../lib/api";
import type { ChatMessage, SearchResult } from "../lib/types";

const SESSION_STORAGE_KEY = "notion-rag.chat-history";

export function useChat() {
  const [question, setQuestion] = useState("");
  const [queryStatus, setQueryStatus] = useState("질문을 입력하면 관련 문서를 찾아 답변합니다.");
  const [asking, setAsking] = useState(false);
  const [history, setHistory] = useState<ChatMessage[]>(() => loadHistory());
  const [selectedSource, setSelectedSource] = useState<SearchResult | null>(null);

  const lastAssistantMessage = useMemo(
    () => [...history].reverse().find((entry) => entry.role === "assistant") ?? null,
    [history],
  );

  useEffect(() => {
    window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(history));
  }, [history]);

  async function submitQuestion(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const trimmed = question.trim();
    if (!trimmed) {
      setQueryStatus("질문을 입력하세요.");
      return;
    }

    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: "user",
      text: trimmed,
      createdAt: new Date().toISOString(),
    };

    setAsking(true);
    setQueryStatus("관련 문서를 검색하고 답변을 생성하는 중");
    setHistory((current) => [...current, userMessage]);
    setQuestion("");

    try {
      const payload = await queryKnowledge(trimmed);
      const assistantMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: "assistant",
        text: payload.answer,
        createdAt: new Date().toISOString(),
        results: payload.results,
      };

      setHistory((current) => [...current, assistantMessage]);
      setSelectedSource(payload.results[0] ?? null);
      setQueryStatus(`${payload.results.length}개의 관련 문서를 기반으로 답변 생성 완료`);
    } catch (error) {
      const message = error instanceof Error ? error.message : "질의 실패";
      setQueryStatus(message);
      setHistory((current) => [
        ...current,
        {
          id: crypto.randomUUID(),
          role: "assistant",
          text: message,
          createdAt: new Date().toISOString(),
        },
      ]);
    } finally {
      setAsking(false);
    }
  }

  function clearHistory() {
    setHistory([]);
    setSelectedSource(null);
    window.localStorage.removeItem(SESSION_STORAGE_KEY);
  }

  return {
    question,
    setQuestion,
    queryStatus,
    asking,
    history,
    selectedSource,
    setSelectedSource,
    lastAssistantMessage,
    visibleResults: lastAssistantMessage?.results ?? [],
    submitQuestion,
    clearHistory,
  };
}

function loadHistory(): ChatMessage[] {
  if (typeof window === "undefined") {
    return [];
  }

  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY);
  if (!raw) {
    return [];
  }

  try {
    const parsed = JSON.parse(raw) as ChatMessage[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}
