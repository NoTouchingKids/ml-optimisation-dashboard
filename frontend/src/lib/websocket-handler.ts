import React, { useEffect, useState } from "react";
import { modelApi } from "../lib/model-api";

interface Log {
  timestamp: number;
  message: string;
  level: string;
}

export function useModelWebSocket() {
  const [logs, setLogs] = useState<Log[]>([]);
  const [status, setStatus] = useState<string>("idle");

  useEffect(() => {
    modelApi.connectWebSocket((message) => {
      switch (message.type) {
        case "live_log":
          setLogs((prev) => [...prev, message.payload]);
          break;
        case "model_status":
          setStatus(message.payload.status);
          break;
      }
    });
  }, []);

  return {
    logs,
    status,
    clearLogs: () => setLogs([]),
  };
}
