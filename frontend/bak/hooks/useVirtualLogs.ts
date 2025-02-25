import { useVirtualizer } from "@tanstack/react-virtual";
import { useRef } from "react";
import type { LogEntry } from "@/types/process";

export function useVirtualLogs(logs: LogEntry[]) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: logs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 80, // Estimated height of each log entry
    overscan: 5, // Number of items to render outside of the visible area
  });

  return { parentRef, virtualizer };
}
