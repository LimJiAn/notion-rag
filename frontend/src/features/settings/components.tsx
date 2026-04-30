import type { Dispatch, ReactNode, SetStateAction } from "react";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import type { SettingsForm, SettingsResponse } from "../../lib/types";
import { cn } from "../../lib/utils";

export function SettingsLayout({
  settingsForm,
  setSettingsForm,
  settingsMeta,
  settingsStatus,
  savingSettings,
  syncing,
  onSave,
  onSaveAndSync,
}: {
  settingsForm: SettingsForm;
  setSettingsForm: Dispatch<SetStateAction<SettingsForm>>;
  settingsMeta: SettingsResponse | null;
  settingsStatus: string;
  savingSettings: boolean;
  syncing: boolean;
  onSave: () => Promise<void>;
  onSaveAndSync: () => Promise<void>;
}) {
  return (
    <main className="grid gap-4 xl:grid-cols-[minmax(0,1.15fr)_360px]">
      <Card className="p-4">
        <CardHeader className="border-b border-ink-100 pb-3">
          <CardTitle>API 설정</CardTitle>
          <CardDescription>Backend data volume</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 p-0 pt-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="Notion Token">
              <Input
                type="password"
                placeholder={settingsMeta?.notion_token_set ? "이미 저장됨" : "secret_xxx"}
                value={settingsForm.notionToken}
                onChange={(event) =>
                  setSettingsForm((current) => ({ ...current, notionToken: event.target.value }))
                }
              />
            </Field>
            <Field label="Gemini API Key">
              <Input
                type="password"
                placeholder={settingsMeta?.gemini_api_key_set ? "이미 저장됨" : "AIza..."}
                value={settingsForm.geminiAPIKey}
                onChange={(event) =>
                  setSettingsForm((current) => ({ ...current, geminiAPIKey: event.target.value }))
                }
              />
            </Field>
            <Field className="md:col-span-2" label="Notion Root Page IDs">
              <Input
                type="text"
                placeholder="uuid1,uuid2"
                value={settingsForm.notionRootPageIDs}
                onChange={(event) =>
                  setSettingsForm((current) => ({ ...current, notionRootPageIDs: event.target.value }))
                }
              />
            </Field>
            <Field label="Notion Version">
              <Input
                type="text"
                value={settingsForm.notionVersion}
                onChange={(event) =>
                  setSettingsForm((current) => ({ ...current, notionVersion: event.target.value }))
                }
              />
            </Field>
            <Field label="Embedding Model">
              <Input
                type="text"
                value={settingsForm.embeddingModel}
                onChange={(event) =>
                  setSettingsForm((current) => ({ ...current, embeddingModel: event.target.value }))
                }
              />
            </Field>
            <Field label="Generation Model">
              <Input
                type="text"
                value={settingsForm.generationModel}
                onChange={(event) =>
                  setSettingsForm((current) => ({ ...current, generationModel: event.target.value }))
                }
              />
            </Field>
          </div>

          <div className="flex flex-wrap gap-2 border-t border-ink-100 pt-4">
            <Button disabled={savingSettings} onClick={() => void onSave()} type="button">
              {savingSettings ? "저장 중..." : "설정 저장"}
            </Button>
            <Button
              variant="secondary"
              disabled={savingSettings || syncing}
              onClick={() => void onSaveAndSync()}
              type="button"
            >
              {savingSettings || syncing ? "처리 중..." : "저장 후 동기화"}
            </Button>
          </div>

          <p className="rounded-md bg-ink-50 px-3 py-2 text-sm text-ink-500">{settingsStatus}</p>
        </CardContent>
      </Card>

      <Card className="p-4">
        <CardHeader className="border-b border-ink-100 pb-3">
          <CardTitle>설정 상태</CardTitle>
          <CardDescription>Runtime values</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 p-0 pt-4">
          <StatusItem label="Notion Token" value={settingsMeta?.notion_token_set ? "Configured" : "Missing"} />
          <StatusItem
            label="Gemini API Key"
            value={settingsMeta?.gemini_api_key_set ? "Configured" : "Missing"}
          />
          <StatusItem
            label="Root Pages"
            value={`${settingsMeta?.notion_root_page_ids.length ?? 0} entries`}
          />
          <StatusItem label="Embedding Model" value={settingsMeta?.embedding_model ?? "-"} />
          <StatusItem label="Generation Model" value={settingsMeta?.generation_model ?? "-"} />
        </CardContent>
      </Card>
    </main>
  );
}

function Field({
  label,
  children,
  className,
}: {
  label: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <label className={cn("space-y-2", className)}>
      <span className="text-sm font-medium text-ink-700">{label}</span>
      {children}
    </label>
  );
}

function StatusItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-md border border-ink-100 bg-white px-3 py-3">
      <span className="text-sm text-ink-500">{label}</span>
      <span className="truncate text-sm font-semibold text-ink-900">{value}</span>
    </div>
  );
}
