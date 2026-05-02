import { Badge } from "./components/ui/badge";
import { Button } from "./components/ui/button";
import { AssistantLayout } from "./features/assistant/components";
import { KnowledgeLayout } from "./features/knowledge/components";
import { SettingsLayout } from "./features/settings/components";
import { formatDate } from "./lib/format";
import { cn } from "./lib/utils";
import { useChat } from "./hooks/use-chat";
import { useDashboard } from "./hooks/use-dashboard";
import { useKnowledge } from "./hooks/use-knowledge";
import { useSettings } from "./hooks/use-settings";
import { useState } from "react";

type ActiveTab = "assistant" | "knowledge" | "settings";

export function App() {
  const [activeTab, setActiveTab] = useState<ActiveTab>("assistant");
  const dashboard = useDashboard();
  const chat = useChat();
  const knowledge = useKnowledge();
  const settings = useSettings(handleSync);

  async function handleSync() {
    await dashboard.handleSync();
    await knowledge.refreshDocuments();
  }

  return (
    <div className="app-bg min-h-screen">
      <div className="app-layout mx-auto grid min-h-screen w-full max-w-[1500px] gap-4 px-4 py-4 sm:px-6 lg:px-8">
        <aside className="surface subtle-grid flex flex-col justify-between overflow-hidden p-4">
          <div className="space-y-6">
            <div>
              <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-md bg-ink-900 text-sm font-bold text-white">
                NR
              </div>
              <h1 className="text-xl font-semibold text-ink-900">Notion RAG</h1>
              <p className="mt-2 text-sm leading-6 text-ink-600">
                Notion 기록을 수집하고 Weaviate 검색으로 답변하는 개인 지식 콘솔
              </p>
            </div>

            <HealthBadge health={dashboard.health} text={dashboard.healthText} />

            <nav className="space-y-2">
              <TabButton active={activeTab === "assistant"} onClick={() => setActiveTab("assistant")}>
                Assistant
              </TabButton>
              <TabButton active={activeTab === "knowledge"} onClick={() => setActiveTab("knowledge")}>
                Knowledge
              </TabButton>
              <TabButton active={activeTab === "settings"} onClick={() => setActiveTab("settings")}>
                Settings
              </TabButton>
            </nav>

            <div className="rounded-lg border border-ink-100 bg-white/80 p-3">
              <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-ink-500">Pipeline</p>
              <div className="space-y-2 text-sm text-ink-700">
                <PipelineStep index="01" label="Notion crawl" />
                <PipelineStep index="02" label="Gemini embedding" />
                <PipelineStep index="03" label="Weaviate search" />
                <PipelineStep index="04" label="Grounded answer" />
              </div>
            </div>
          </div>

          <div className="mt-6 rounded-lg border border-ink-100 bg-ink-900 p-3 text-white">
            <p className="text-xs uppercase tracking-wide text-white/60">Local stack</p>
            <p className="mt-2 text-sm font-medium">Backend 8080 · Frontend 3000 · Weaviate 8081</p>
          </div>
        </aside>

        <div className="flex min-w-0 flex-col gap-4">
          <header className="surface flex flex-col gap-3 p-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase tracking-wide text-ember-500">
                {activeTabLabel(activeTab)}
              </p>
              <h2 className="mt-1 text-2xl font-semibold text-ink-900">{activeTabTitle(activeTab)}</h2>
            </div>

            <div className="flex flex-wrap gap-2 sm:items-center">
              <Metric label="청크" value={String(dashboard.stats.documents ?? "-")} />
              <Metric label="저장소" value={dashboard.stats.vector_store ?? "file"} />
              <Metric label="최근 동기화" value={formatDate(dashboard.stats.last_updated)} />
              <Button variant="secondary" onClick={handleSync} disabled={dashboard.syncing}>
                {dashboard.syncing ? "동기화 중" : "동기화"}
              </Button>
              <Button variant="outline" onClick={chat.clearHistory} disabled={chat.history.length === 0}>
                대화 초기화
              </Button>
            </div>
          </header>

          {activeTab === "assistant" && (
            <AssistantLayout
              asking={chat.asking}
              history={chat.history}
              lastAssistantMessage={chat.lastAssistantMessage}
              onQuestionChange={chat.setQuestion}
              onQuestionSubmit={chat.submitQuestion}
              onSourceSelect={chat.setSelectedSource}
              onSuggestedQuestionSelect={chat.applySuggestedQuestion}
              queryStatus={chat.queryStatus}
              question={chat.question}
              selectedSource={chat.selectedSource}
              visibleResults={chat.visibleResults}
            />
          )}

          {activeTab === "knowledge" && (
            <KnowledgeLayout
              documents={knowledge.documents}
              loading={knowledge.loading}
              onRefresh={knowledge.refreshDocuments}
              stats={dashboard.stats}
              status={knowledge.status}
            />
          )}

          {activeTab === "settings" && (
            <SettingsLayout
              onSave={() => settings.saveSettings(false)}
              onSaveAndSync={() => settings.saveSettings(true)}
              savingSettings={settings.savingSettings}
              settingsForm={settings.settingsForm}
              settingsMeta={settings.settingsMeta}
              setSettingsForm={settings.setSettingsForm}
              settingsStatus={settings.settingsStatus}
              syncing={dashboard.syncing}
            />
          )}
        </div>
      </div>
    </div>
  );
}

function HealthBadge({ health, text }: { health: "loading" | "online" | "offline"; text: string }) {
  return (
    <Badge
      className={cn(
        "gap-1.5",
        health === "online" && "bg-moss-100 text-moss-700",
        health === "offline" && "bg-red-50 text-red-700",
        health === "loading" && "bg-ink-50 text-ink-500",
      )}
    >
      <span
        className={cn(
          "h-1.5 w-1.5 rounded-full",
          health === "online" && "bg-moss-500",
          health === "offline" && "bg-red-500",
          health === "loading" && "bg-ink-300",
        )}
      />
      {text}
    </Badge>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-[108px] rounded-md border border-ink-100 bg-ink-50 px-3 py-2">
      <span className="block text-xs text-ink-500">{label}</span>
      <span className="mt-1 block truncate text-sm font-semibold text-ink-900">{value}</span>
    </div>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: string;
}) {
  return (
    <button
      className={cn(
        "flex w-full items-center rounded-md px-3 py-2 text-sm font-medium transition",
        active
          ? "bg-ink-900 text-white"
          : "text-ink-500 hover:bg-ink-50 hover:text-ink-900",
      )}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}

function PipelineStep({ index, label }: { index: string; label: string }) {
  return (
    <div className="flex items-center gap-2">
      <span className="flex h-6 w-6 items-center justify-center rounded bg-ember-100 text-[11px] font-semibold text-ember-700">
        {index}
      </span>
      <span>{label}</span>
    </div>
  );
}

function activeTabLabel(tab: ActiveTab) {
  switch (tab) {
    case "assistant":
      return "Ask";
    case "knowledge":
      return "Inspect";
    case "settings":
      return "Configure";
  }
}

function activeTabTitle(tab: ActiveTab) {
  switch (tab) {
    case "assistant":
      return "근거 기반 질문 응답";
    case "knowledge":
      return "벡터 인덱스 상태";
    case "settings":
      return "API 및 동기화 설정";
  }
}
