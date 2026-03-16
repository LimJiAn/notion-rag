import type { QueryResponse, SettingsResponse, Stats } from "./types";

export async function getHealth() {
  const response = await fetch("/healthz");
  if (!response.ok) {
    throw new Error("health check failed");
  }
  return (await response.json()) as { status: string };
}

export async function getStats() {
  const response = await fetch("/api/v1/stats");
  if (!response.ok) {
    throw new Error("stats request failed");
  }
  return (await response.json()) as Stats;
}

export async function syncNotion() {
  const response = await fetch("/api/v1/sync", { method: "POST" });
  const payload = (await response.json()) as { error?: string; stats?: { chunks?: number } };
  if (!response.ok) {
    throw new Error(payload.error || "sync failed");
  }
  return payload;
}

export async function queryKnowledge(question: string) {
  const response = await fetch("/api/v1/query", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ question }),
  });
  const payload = (await response.json()) as QueryResponse & { error?: string };
  if (!response.ok) {
    throw new Error(payload.error || "query failed");
  }
  return payload;
}

export async function getSettings() {
  const response = await fetch("/api/v1/settings");
  if (!response.ok) {
    throw new Error("settings request failed");
  }
  return (await response.json()) as SettingsResponse;
}

export async function saveSettings(input: {
  notion_token: string;
  notion_version: string;
  notion_root_page_ids: string;
  gemini_api_key: string;
  embedding_model: string;
  generation_model: string;
}) {
  const response = await fetch("/api/v1/settings", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  const payload = (await response.json()) as { error?: string; settings?: SettingsResponse };
  if (!response.ok) {
    throw new Error(payload.error || "settings save failed");
  }
  return payload;
}
