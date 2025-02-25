import { useState, useEffect } from "react";
import type { LogEntry, LogMetrics } from "@/types/process";

export function useProcessMetrics(logs: LogEntry[]) {
  const [metrics, setMetrics] = useState<LogMetrics>({
    totalLogs: 0,
    errorCount: 0,
    warningCount: 0,
    startTime: null,
    status: "idle",
  });

  useEffect(() => {
    const errorCount = logs.filter((log) => log.level === "ERROR").length;
    const warningCount = logs.filter((log) => log.level === "WARNING").length;

    setMetrics((prev) => ({
      ...prev,
      totalLogs: logs.length,
      errorCount,
      warningCount,
    }));
  }, [logs]);

  const startProcess = () => {
    setMetrics((prev) => ({
      ...prev,
      startTime: new Date(),
      status: "running",
    }));
  };

  const completeProcess = () => {
    setMetrics((prev) => ({
      ...prev,
      status: "completed",
    }));
  };

  const errorProcess = () => {
    setMetrics((prev) => ({
      ...prev,
      status: "error",
    }));
  };

  return { metrics, startProcess, completeProcess, errorProcess };
}
