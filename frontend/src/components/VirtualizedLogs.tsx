import React, { useEffect, useState } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import type { LogRecord, LogRecordPayload } from "@/types/log";
import { base64ToUint8Array, decompressData } from "@/lib/utils";
import * as msgpack from "msgpack-lite";
import { LogEntry } from "./LogEntry";

interface VirtualizedLogsProps {
  logs?: LogRecordPayload[];
}

const UUID_LENGTH = 36; // Standard UUID string length

const processLog = (raw: LogRecordPayload): LogRecord | undefined => {
  try {
    const logRecord = raw;
    const compressedData = base64ToUint8Array(logRecord.message);
    const decompressedData = decompressData(compressedData.slice(UUID_LENGTH));
    return msgpack.decode(decompressedData) as LogRecord;
  } catch (error) {
    console.error("Error processing log:", error);
  }
};

export function VirtualizedLogs({ logs = [] }: VirtualizedLogsProps) {
  const parentRef = React.useRef<HTMLDivElement>(null);
  const innerRef = React.useRef<HTMLDivElement>(null);
  const rowRefsMap = React.useRef(new Map<number, HTMLDivElement>());

  const virtualizer = useVirtualizer({
    count: logs?.length ?? 0,
    getScrollElement: () => parentRef.current,
    // Update the estimateSize to dynamically calculate item height
    estimateSize: () => 50,
    overscan: 5,
    measureElement: (el) => el.getBoundingClientRect().height,
    onChange: (instance) => {
      // innerRef.current!.style.height = `${instance.getTotalSize()}px`;
      instance.getVirtualItems().forEach((virtualRow) => {
        const rowRef = rowRefsMap.current.get(virtualRow.index);
        if (!rowRef) return;
        rowRef.style.transform = `translateY(${virtualRow.start}px)`;
      });
    },
  });

  useEffect(() => {
    if (parentRef.current) {
      parentRef.current.scrollTop = parentRef.current.scrollHeight;
    }
  }, [logs?.length]);

  return (
    <div
      ref={parentRef}
      className="h-full overflow-auto"
      style={{ contain: "strict" }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: "100%",
          position: "relative",
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => {
          const log_raw = logs?.[virtualItem.index];
          const log = processLog(log_raw);
          if (!log || log === null) return null;

          return (
            <div
              ref={innerRef}
              style={{
                width: "100%",
                position: "relative",
              }}
            >
              <LogEntry
                key={virtualItem.key}
                log={log}
                measureRef={(el) => virtualizer.measureElement(el)}
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  transform: `translateY(${virtualItem.start}px)`,
                }}
                data-index={virtualItem.index} // Now this will be added to the outer <div>
              />
            </div>
          );

          // return (
          //   <div
          //     key={`${log.time}-${log.process}`}
          //     ref={(el) => (itemRefs.current[virtualItem.index] = el)} // assign ref to each item
          //     style={{
          //       position: "absolute",
          //       top: 0,
          //       left: 0,
          //       width: "100%",
          //       height: `${virtualItem.size}px`,
          //       transform: `translateY(${virtualItem.start}px)`,
          //     }}
          //     className="border-b py-1 px-2 text-sm font-mono"
          //   >
          //     <span className="text-muted-foreground">
          //       {new Date(log.time).toISOString()}
          //     </span>{" "}
          //     <span className={getLevelStyle(log.levelname)}>
          //       [{log.levelname}]
          //     </span>{" "}
          //     <span>[{log.processName}]</span>{" "}
          //     <span>{log.message || log.msg}</span>
          //   </div>
          // );
        })}
      </div>
    </div>
  );
}
