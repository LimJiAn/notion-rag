import { useEffect, useState } from "react";

import { getDocuments } from "../lib/api";
import type { KnowledgeDocument } from "../lib/types";

export function useKnowledge() {
  const [documents, setDocuments] = useState<KnowledgeDocument[]>([]);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState("인덱스 문서를 불러오지 않았습니다.");

  useEffect(() => {
    void refreshDocuments();
  }, []);

  async function refreshDocuments() {
    setLoading(true);
    setStatus("인덱스 문서를 불러오는 중입니다.");

    try {
      const payload = await getDocuments(25);
      setDocuments(payload.documents);
      setStatus(`${payload.count}개 청크를 불러왔습니다.`);
    } catch (error) {
      setDocuments([]);
      setStatus(error instanceof Error ? error.message : "인덱스 문서를 불러오지 못했습니다.");
    } finally {
      setLoading(false);
    }
  }

  return {
    documents,
    loading,
    status,
    refreshDocuments,
  };
}
