export interface LogMetrics {
  totalLogs: number;
  errorCount: number;
  warningCount: number;
  startTime: Date | null;
  status: "idle" | "running" | "error" | "completed";
}

export interface LogEntry {
  id: string;
  timestamp: number;
  level: "INFO" | "ERROR" | "WARNING" | "SUCCESS";
  message: string;
  details?: Record<string, unknown>;
  clientId: string;
}
