import type { ModelAPI } from "@/lib/model-api";
import type { LogRecordPayload } from "@/types/log";
import { useState, useEffect } from "react";

export function useLogStream(api: ModelAPI, isRunning: boolean) {
  const [logs, setLogs] = useState<LogRecordPayload[]>([]);

  useEffect(() => {
    let unsubscribe: (() => void) | undefined;

    if (isRunning) {
      unsubscribe = api.connectWebSocket<LogRecordPayload>(
        "live_log",
        (log) => {
          setLogs((prev) => [...prev, log]);
        }
      );
    }

    return () => {
      unsubscribe?.();
      if (!isRunning) {
        setLogs([]);
      }
    };
  }, [api, isRunning]);

  return { logs, clearLogs: () => setLogs([]) };
}
