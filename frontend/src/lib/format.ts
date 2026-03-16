export function trimSnippet(value: string) {
  const clean = value.replace(/\s+/g, " ").trim();
  if (clean.length <= 240) {
    return clean;
  }
  return `${clean.slice(0, 240)}...`;
}

export function buildNotionPageURL(pageID: string) {
  const normalized = pageID.replace(/-/g, "").trim();
  if (!normalized) {
    return null;
  }
  return `https://www.notion.so/${normalized}`;
}

export function openNotionPage(pageID: string) {
  const url = buildNotionPageURL(pageID);
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
