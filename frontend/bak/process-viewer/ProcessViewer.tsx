import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { MetricsCard } from "./MetricsCard";
import { ClientIdInput } from "./ClientIdInput";
import { LogEntry } from "./LogEntry";
import type { LogRecord } from "@/types/log";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useVirtualizer } from "@tanstack/react-virtual";
import { base64ToUint8Array, decompressData } from "@/lib/utils";
import * as msgpack from "msgpack-lite";

const UUID_LENGTH = 36; // Standard UUID string length

export default function ProcessViewer() {
  const [logs, setLogs] = useState<LogRecord[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [clientId, setClientId] = useState("");
  const [startTime, setStartTime] = useState<Date | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const parentRef = useRef<HTMLDivElement>(null);

  // Virtual scrolling setup
  const rowVirtualizer = useVirtualizer({
    count: logs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 80,
    overscan: 5,
  });

  // Metrics calculations
  const totalLogs = logs.length;
  const errorCount = logs.filter((log) => log.levelname === "ERROR").length;
  const errorRate = totalLogs
    ? ((errorCount / totalLogs) * 100).toFixed(1)
    : "0";
  const runtime = startTime
    ? Math.round((new Date().getTime() - startTime.getTime()) / 1000)
    : 0;
  const processStatus = isConnected ? "Running" : "Disconnected";

  const startProcess = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/model/train", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          client_id: clientId,
          train_start_date: "2024-01-01",
          train_end_date: "2024-02-01",
          model_config: {
            feature_engineering: true,
            detrend: false,
            difference: false,
          },
        }),
      });

      if (response.ok) {
        connectWebSocket();
        setStartTime(new Date());
      }
    } catch (error) {
      console.error("Error starting process:", error);
    }
  };

  const connectWebSocket = () => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = new WebSocket(`ws://localhost:8080/ws?clientId=${clientId}`);
    wsRef.current = ws;

    ws.onopen = () => setIsConnected(true);
    ws.onclose = () => setIsConnected(false);
    ws.onmessage = handleWebSocketMessage;
  };

  const handleWebSocketMessage = async (event: MessageEvent) => {
    try {
      const logRecord = JSON.parse(event.data).payload;
      const compressedData = await base64ToUint8Array(logRecord.message);
      const decompressedData = await decompressData(
        compressedData.slice(UUID_LENGTH)
      );
      const log = msgpack.decode(decompressedData);
      setLogs((prev) => [
        ...prev,
        {
          ...log,
          time: new Date(logRecord.timestamp / 1000000),
          clientId: logRecord.clientId,
        },
      ]);
    } catch (error) {
      console.error("Error processing WebSocket message:", error);
    }
  };

  const clearLogs = () => {
    setLogs([]);
    setStartTime(null);
  };

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <MetricsCard title="Total Logs" value={totalLogs} />
        <MetricsCard title="Error Rate" value={`${errorRate}%`} />
        <MetricsCard title="Process Status" value={processStatus} />
        <MetricsCard title="Runtime" value={`${runtime}s`} />
      </div>

      <div className="flex items-center space-x-4">
        <div className="flex-1">
          <ClientIdInput onIdChange={setClientId} disabled={isConnected} />
        </div>
        <Button onClick={startProcess} disabled={isConnected || !clientId}>
          Start Process
        </Button>
        <Button variant="outline" onClick={clearLogs}>
          Clear Logs
        </Button>
      </div>

      <div
        ref={parentRef}
        className="h-[1000px] overflow-auto border rounded-lg bg-background"
      >
        <div
          style={{
            height: `${rowVirtualizer.getTotalSize()}px`,
            width: "100%",
            position: "relative",
          }}
        >
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const log = logs[virtualRow.index];
            return (
              <LogEntry
                key={virtualRow.index}
                log={log}
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  transform: `translateY(${virtualRow.start}px)`,
                }}
              />
            );
          })}
        </div>
      </div>
    </div>
  );
}
