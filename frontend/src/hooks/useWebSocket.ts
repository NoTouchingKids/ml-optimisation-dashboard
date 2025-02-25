import { useState, useEffect } from "react";
import type { ModelAPI } from "../lib/model-api";
import type { LogRecordPayload } from "@/types/log";

export function useWebSocket(api: ModelAPI, messageType: string) {
  const [logs, setLogs] = useState<LogRecordPayload[]>([]);

  useEffect(() => {
    const unsubscribe = api.connectWebSocket<LogRecordPayload>(
      messageType,
      (data) => {
        setLogs((prev) => [...prev, data]);
      }
    );

    return unsubscribe;
  }, [api, messageType]);

  return { logs };
}
