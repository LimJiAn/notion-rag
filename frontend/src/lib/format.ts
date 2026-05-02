export function trimSnippet(value: string, max = 240) {
  const clean = value.replace(/\s+/g, " ").trim();
  if (clean.length <= max) {
    return clean;
  }
  return `${clean.slice(0, max)}...`;
}

export function buildNotionPageURL(pageID: string) {
  const normalized = pageID.replace(/-/g, "").trim();
  if (!normalized) {
    return null;
  }
  return `https://www.notion.so/${normalized}`;
}

export function resolveNotionURL(pageID: string, notionURL?: string) {
  return notionURL || buildNotionPageURL(pageID);
}

export function openNotionPage(pageID: string, notionURL?: string) {
  const url = resolveNotionURL(pageID, notionURL);
  if (!url) {
    return;
  }
  window.open(url, "_blank", "noopener,noreferrer");
}

export function formatDate(value?: string) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ko-KR", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function formatConfidence(label?: "high" | "medium" | "low", score?: number) {
  const percent = score ? Math.round(score * 100) : 0;

  switch (label) {
    case "high":
      return `신뢰도 높음 ${percent}%`;
    case "medium":
      return `신뢰도 보통 ${percent}%`;
    case "low":
      return `신뢰도 낮음 ${percent}%`;
    default:
      return "신뢰도 정보 없음";
  }
}
