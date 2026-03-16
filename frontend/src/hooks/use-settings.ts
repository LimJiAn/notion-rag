import { useEffect, useState } from "react";

import { getSettings, saveSettings as persistSettings } from "../lib/api";
import type { SettingsForm, SettingsResponse } from "../lib/types";

const DEFAULT_SETTINGS_FORM: SettingsForm = {
  notionToken: "",
  notionVersion: "2026-03-11",
  notionRootPageIDs: "",
  geminiAPIKey: "",
  embeddingModel: "gemini-embedding-001",
  generationModel: "gemini-2.5-flash",
};

export function useSettings(onSync?: () => Promise<void>) {
  const [settingsForm, setSettingsForm] = useState<SettingsForm>(DEFAULT_SETTINGS_FORM);
  const [settingsMeta, setSettingsMeta] = useState<SettingsResponse | null>(null);
  const [settingsStatus, setSettingsStatus] = useState("설정은 백엔드 데이터 볼륨에 저장됩니다.");
  const [savingSettings, setSavingSettings] = useState(false);

  useEffect(() => {
    void fetchSettings();
  }, []);

  async function fetchSettings() {
    try {
      const payload = await getSettings();
      setSettingsMeta(payload);
      setSettingsForm({
        notionToken: "",
        notionVersion: payload.notion_version,
        notionRootPageIDs: payload.notion_root_page_ids.join(", "),
        geminiAPIKey: "",
        embeddingModel: payload.embedding_model,
        generationModel: payload.generation_model,
      });
    } catch {
      setSettingsStatus("설정 정보를 불러오지 못했습니다.");
    }
  }

  async function saveSettings(syncAfterSave: boolean) {
    setSavingSettings(true);
    setSettingsStatus("설정을 저장하는 중입니다.");

    try {
      const payload = await persistSettings({
        notion_token: settingsForm.notionToken,
        notion_version: settingsForm.notionVersion,
        notion_root_page_ids: settingsForm.notionRootPageIDs,
        gemini_api_key: settingsForm.geminiAPIKey,
        embedding_model: settingsForm.embeddingModel,
        generation_model: settingsForm.generationModel,
      });

      if (payload.settings) {
        setSettingsMeta(payload.settings);
        setSettingsForm((current) => ({
          ...current,
          notionToken: "",
          geminiAPIKey: "",
          notionRootPageIDs:
            payload.settings?.notion_root_page_ids.join(", ") || current.notionRootPageIDs,
        }));
      }

      setSettingsStatus(syncAfterSave ? "설정을 저장했습니다. 이어서 동기화합니다." : "설정을 저장했습니다.");
      if (syncAfterSave && onSync) {
        await onSync();
      }
    } catch (error) {
      setSettingsStatus(error instanceof Error ? error.message : "설정 저장 실패");
    } finally {
      setSavingSettings(false);
    }
  }

  return {
    settingsForm,
    setSettingsForm,
    settingsMeta,
    settingsStatus,
    savingSettings,
    saveSettings,
  };
}
