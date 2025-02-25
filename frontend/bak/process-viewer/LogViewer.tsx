import { useVirtualLogs } from "@/hooks/useVirtualLogs";
import type { LogEntry } from "@/types/process";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

interface LogViewerProps {
  logs: LogEntry[];
}

export function LogViewer({ logs }: LogViewerProps) {
  const { parentRef, virtualizer } = useVirtualLogs(logs);

  return (
    <ScrollArea ref={parentRef} className="h-[600px] rounded-md border">
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: "100%",
          position: "relative",
        }}
      >
        {virtualizer.getVirtualItems().map((virtualRow) => {
          const log = logs[virtualRow.index];
          return (
            <div
              key={virtualRow.key}
              data-index={virtualRow.index}
              ref={virtualizer.measureElement}
              className={cn(
                "absolute top-0 left-0 w-full",
                "border-l-4 p-4 transition-colors",
                {
                  "border-blue-500 bg-blue-50 dark:bg-blue-950":
                    log.level === "INFO",
                  "border-red-500 bg-red-50 dark:bg-red-950":
                    log.level === "ERROR",
                  "border-yellow-500 bg-yellow-50 dark:bg-yellow-950":
                    log.level === "WARNING",
                  "border-green-500 bg-green-50 dark:bg-green-950":
                    log.level === "SUCCESS",
                }
              )}
              style={{
                transform: `translateY(${virtualRow.start}px)`,
              }}
            >
              <div className="flex items-start gap-2">
                <time className="text-sm text-muted-foreground">
                  {new Date(log.timestamp).toLocaleTimeString()}
                </time>
                <span className="font-medium">{log.level}</span>
                <span className="flex-1">{log.message}</span>
              </div>
              {log.details && (
                <div className="mt-2 text-sm text-muted-foreground">
                  {Object.entries(log.details)
                    .map(([key, value]) => `${key}: ${value}`)
                    .join(" | ")}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </ScrollArea>
  );
}
