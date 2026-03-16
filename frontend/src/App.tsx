import { Badge } from "./components/ui/badge";
import { Button } from "./components/ui/button";
import { Card, CardContent } from "./components/ui/card";
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
    <div className="min-h-screen">
      <div className="mx-auto flex min-h-screen w-full max-w-7xl flex-col px-4 py-6 sm:px-6 lg:px-8">
        <header className="surface subtle-grid mb-5 overflow-hidden p-6 sm:p-8">
          <div className="grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_360px]">
            <div className="space-y-5">
              <div className="space-y-3">
                <p className="font-display text-xs font-semibold uppercase tracking-[0.28em] text-ember-500">
                  Personal Knowledge Base
                </p>
                <p className="max-w-2xl text-balance text-lg leading-8 text-ink-700 sm:text-xl">
                  노션에 쌓인 업무 기록, 아이디어, 회의 메모를 동기화하고 바로 질문하세요.
                  답변의 근거가 된 문서와 대화 기록도 함께 확인할 수 있습니다.
                </p>
              </div>

              <div className="flex flex-wrap gap-2">
                <Badge className="bg-white/80 text-ink-700">Linked with Notion</Badge>
                <Badge className="bg-white/80 text-ink-700">Session Memory</Badge>
                <Badge className="bg-white/80 text-ink-700">Grounded Answers</Badge>
              </div>
            </div>

            <Card className="border-white/80 bg-white/72 p-5">
              <CardContent className="flex h-full flex-col justify-between gap-5 p-0">
                <div className="inline-flex w-fit items-center gap-2 rounded-full border border-white/80 bg-white/80 px-3 py-2 text-sm text-ink-700">
                  <span
                    className={cn(
                      "h-2.5 w-2.5 rounded-full",
                      dashboard.health === "online" &&
                        "bg-moss-500 shadow-[0_0_0_6px_rgba(82,114,95,0.14)]",
                      dashboard.health === "offline" &&
                        "bg-ember-500 shadow-[0_0_0_6px_rgba(198,100,59,0.14)]",
                      dashboard.health === "loading" &&
                        "bg-ink-300 shadow-[0_0_0_6px_rgba(201,193,176,0.24)]",
                    )}
                  />
                  {dashboard.healthText}
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="rounded-3xl bg-white/80 p-4">
                    <p className="text-xs uppercase tracking-[0.2em] text-ink-500">Indexed Chunks</p>
                    <p className="mt-2 font-display text-3xl font-semibold text-ink-900">
                      {String(dashboard.stats.documents ?? "-")}
                    </p>
                  </div>
                  <div className="rounded-3xl bg-white/80 p-4">
                    <p className="text-xs uppercase tracking-[0.2em] text-ink-500">Last Sync</p>
                    <p className="mt-2 font-display text-lg font-semibold text-ink-900">
                      {formatDate(dashboard.stats.last_updated)}
                    </p>
                  </div>
                </div>

                <div className="flex flex-wrap gap-3">
                  <Button
                    variant="secondary"
                    onClick={dashboard.handleSync}
                    disabled={dashboard.syncing}
                    className="flex-1"
                  >
                    {dashboard.syncing ? "동기화 중..." : "Notion 동기화"}
                  </Button>
                  <Button variant="outline" onClick={chat.clearHistory} disabled={chat.history.length === 0}>
                    대화 초기화
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </header>

        <div className="mb-5 flex gap-2">
          <TabButton active={activeTab === "assistant"} onClick={() => setActiveTab("assistant")}>
            Assistant
          </TabButton>
          <TabButton active={activeTab === "settings"} onClick={() => setActiveTab("settings")}>
            Settings
          </TabButton>
        </div>

        {activeTab === "assistant" ? (
          <AssistantLayout
            asking={chat.asking}
            history={chat.history}
            lastAssistantMessage={chat.lastAssistantMessage}
            onQuestionChange={chat.setQuestion}
            onQuestionSubmit={chat.submitQuestion}
            onSourceSelect={chat.setSelectedSource}
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
        "rounded-2xl border px-4 py-3 font-display text-sm font-semibold transition",
        active
          ? "border-ink-900 bg-ink-900 text-white"
          : "border-white/80 bg-white/70 text-ink-500 hover:bg-white",
      )}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}
