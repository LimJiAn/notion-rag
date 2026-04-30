import { Badge } from "./components/ui/badge";
import { Button } from "./components/ui/button";
import { AssistantLayout } from "./features/assistant/components";
import { SettingsLayout } from "./features/settings/components";
import { formatDate } from "./lib/format";
import { cn } from "./lib/utils";
import { useChat } from "./hooks/use-chat";
import { useDashboard } from "./hooks/use-dashboard";
import { useSettings } from "./hooks/use-settings";
import { useState } from "react";

type ActiveTab = "assistant" | "settings";

export function App() {
  const [activeTab, setActiveTab] = useState<ActiveTab>("assistant");
  const dashboard = useDashboard();
  const chat = useChat();
  const settings = useSettings(dashboard.handleSync);

  return (
    <div className="min-h-screen bg-ink-50">
      <div className="mx-auto flex min-h-screen w-full max-w-[1440px] flex-col px-4 py-4 sm:px-6 lg:px-8">
        <header className="mb-4 rounded-lg border border-ink-100 bg-white">
          <div className="flex flex-col gap-3 border-b border-ink-100 px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-lg font-semibold text-ink-900">노션 코파일럿</h1>
                <HealthBadge health={dashboard.health} text={dashboard.healthText} />
              </div>
              <p className="mt-1 text-sm text-ink-500">Notion 기반 개인 지식 검색 및 답변 콘솔</p>
            </div>

            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
              <Metric label="청크" value={String(dashboard.stats.documents ?? "-")} />
              <Metric label="최근 동기화" value={formatDate(dashboard.stats.last_updated)} />
              <Button variant="secondary" onClick={dashboard.handleSync} disabled={dashboard.syncing}>
                {dashboard.syncing ? "동기화 중" : "동기화"}
              </Button>
              <Button variant="outline" onClick={chat.clearHistory} disabled={chat.history.length === 0}>
                대화 초기화
              </Button>
            </div>
          </div>

          <nav className="flex gap-1 px-3 py-2">
            <TabButton active={activeTab === "assistant"} onClick={() => setActiveTab("assistant")}>
              Assistant
            </TabButton>
            <TabButton active={activeTab === "settings"} onClick={() => setActiveTab("settings")}>
              Settings
            </TabButton>
          </nav>
        </header>

        {activeTab === "assistant" ? (
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
        ) : (
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
    <div className="flex min-w-[120px] items-center justify-between gap-3 rounded-md border border-ink-100 bg-ink-50 px-3 py-2">
      <span className="text-xs text-ink-500">{label}</span>
      <span className="truncate text-sm font-semibold text-ink-900">{value}</span>
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
        "rounded-md px-3 py-2 text-sm font-medium transition",
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
