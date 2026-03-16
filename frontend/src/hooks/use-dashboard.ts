import { useEffect, useState } from "react";

import { getHealth, getStats, syncNotion } from "../lib/api";
import type { HealthStatus, Stats } from "../lib/types";

export function useDashboard() {
  const [health, setHealth] = useState<HealthStatus>("loading");
  const [healthText, setHealthText] = useState("서버 상태 확인 중");
  const [stats, setStats] = useState<Stats>({});
  const [syncing, setSyncing] = useState(false);

  useEffect(() => {
    void refreshHealth();
    void refreshStats();
  }, []);

  async function refreshHealth() {
    try {
      await getHealth();
      setHealth("online");
      setHealthText("백엔드 연결 정상");
    } catch {
      setHealth("offline");
      setHealthText("백엔드에 연결할 수 없음");
    }
  }

  async function refreshStats() {
    try {
      setStats(await getStats());
    } catch {
      setStats({});
    }
  }

  async function handleSync() {
    setSyncing(true);
    setHealthText("Notion 동기화 실행 중");

    try {
      const payload = await syncNotion();
      setHealth("online");
      setHealthText(`동기화 완료, ${payload.stats?.chunks ?? 0}개 청크 반영`);
      await refreshStats();
    } catch (error) {
      setHealth("offline");
      setHealthText(error instanceof Error ? error.message : "동기화 실패");
    } finally {
      setSyncing(false);
    }
  }

  return {
    health,
    healthText,
    stats,
    syncing,
    handleSync,
    refreshStats,
  };
}
