import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { formatDate, openNotionPage, trimSnippet } from "../../lib/format";
import type { KnowledgeDocument, Stats } from "../../lib/types";
import { cn } from "../../lib/utils";

export function KnowledgeLayout({
  stats,
  documents,
  loading,
  status,
  onRefresh,
}: {
  stats: Stats;
  documents: KnowledgeDocument[];
  loading: boolean;
  status: string;
  onRefresh: () => Promise<void>;
}) {
  const storeStatus = stats.status ?? (stats.vector_store ? "ready" : "unknown");
  const endpoint = resolveDisplayEndpoint(stats.weaviate_url);

  return (
    <main className="grid flex-1 gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
      <section className="space-y-4">
        <Card className="p-4">
          <CardHeader className="border-b border-ink-100 pb-3">
            <CardTitle>Vector Store</CardTitle>
            <CardDescription>current index backend</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 p-0 pt-4">
            <InfoRow label="store" value={stats.vector_store ?? "file"} />
            <InfoRow label="status" value={storeStatus} tone={storeStatus === "unavailable" ? "danger" : "default"} />
            <InfoRow label="documents" value={String(stats.documents ?? 0)} />
            <InfoRow label="collection" value={stats.collection ?? "-"} />
            <InfoRow label="updated" value={formatDate(stats.last_updated)} />
            {endpoint ? <InfoRow label="endpoint" value={endpoint} /> : null}
            {stats.error ? (
              <div className="rounded-md border border-red-100 bg-red-50 px-3 py-2 text-xs leading-5 text-red-700">
                {stats.error}
              </div>
            ) : null}
          </CardContent>
        </Card>

        <Card className="p-4">
          <CardHeader className="border-b border-ink-100 pb-3">
            <CardTitle>Operations</CardTitle>
            <CardDescription>index inspection</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 p-0 pt-4">
            <Button className="w-full" disabled={loading} onClick={onRefresh} type="button">
              {loading ? "불러오는 중" : "문서 새로고침"}
            </Button>
            <p className="text-sm leading-6 text-ink-500">{status}</p>
          </CardContent>
        </Card>
      </section>

      <Card className="p-4">
        <CardHeader className="flex-row items-center justify-between border-b border-ink-100 pb-3">
          <div>
            <CardTitle>Indexed Chunks</CardTitle>
            <CardDescription>{documents.length} loaded chunks</CardDescription>
          </div>
          <Badge className={cn(stats.vector_store === "weaviate" ? "bg-moss-100 text-moss-700" : "")}>
            {stats.vector_store === "weaviate" ? "Weaviate" : "File"}
          </Badge>
        </CardHeader>
        <CardContent className="p-0 pt-4">
          {documents.length === 0 ? (
            <div className="rounded-md border border-dashed border-ink-200 bg-ink-50 px-4 py-8 text-sm text-ink-500">
              동기화된 청크가 없습니다. 먼저 Notion 동기화를 실행하세요.
            </div>
          ) : (
            <div className="divide-y divide-ink-100 rounded-md border border-ink-100">
              {documents.map((document) => (
                <button
                  className="block w-full bg-white p-3 text-left transition first:rounded-t-md last:rounded-b-md hover:bg-ink-50"
                  key={`${document.page_id}-${document.chunk}-${document.title}`}
                  onDoubleClick={() => openNotionPage(document.page_id, document.notion_url)}
                  type="button"
                >
                  <div className="mb-2 flex items-start justify-between gap-4">
                    <div>
                      <h3 className="font-medium text-ink-900">{document.title || "Untitled"}</h3>
                      <p className="mt-1 text-xs text-ink-500">
                        page_id: {document.page_id} · chunk {document.chunk}
                      </p>
                    </div>
                    <span className="text-xs font-medium text-ember-500">더블클릭 열기</span>
                  </div>
                  {document.updated_at ? (
                    <p className="mb-2 text-xs text-ink-500">updated {formatDate(document.updated_at)}</p>
                  ) : null}
                  <p className="text-sm leading-6 text-ink-700">{trimSnippet(document.content, 260)}</p>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </main>
  );
}

function resolveDisplayEndpoint(value?: string) {
  if (!value) {
    return "";
  }
  return value.replace("http://weaviate:8080", "http://localhost:8081");
}

function InfoRow({
  label,
  value,
  tone = "default",
}: {
  label: string;
  value: string;
  tone?: "default" | "danger";
}) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md border border-ink-100 bg-ink-50 px-3 py-2">
      <span className="text-xs text-ink-500">{label}</span>
      <span className={cn("truncate text-sm font-semibold text-ink-900", tone === "danger" && "text-red-700")}>
        {value || "-"}
      </span>
    </div>
  );
}
