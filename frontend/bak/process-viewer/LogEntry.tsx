// src/components/process-viewer/LogEntry.tsx
import { useState } from "react";
import { ChevronRight, ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { LogRecord } from "@/types/log";

interface LogEntryProps {
  log: LogRecord;
  style?: React.CSSProperties;
}

export function LogEntry({ log, style }: LogEntryProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const getLogColor = (level: string) => {
    switch (level) {
      case "INFO":
        return "border-blue-500 bg-blue-50/50 dark:bg-blue-950/50";
      case "ERROR":
        return "border-red-500 bg-red-50/50 dark:bg-red-950/50";
      case "WARNING":
        return "border-yellow-500 bg-yellow-50/50 dark:bg-yellow-950/50";
      case "SUCCESS":
        return "border-green-500 bg-green-50/50 dark:bg-green-950/50";
      default:
        return "border-gray-500 bg-gray-50/50 dark:bg-gray-950/50";
    }
  };

  const formatTime = (date: string | Date) => {
    if (typeof date === "string") {
      date = new Date(date);
    }
    return date.toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      fractionalSecondDigits: 3,
    });
  };

  return (
    <div
      className={cn(
        "border-l-4 transition-colors mb-1",
        getLogColor(log.levelname)
      )}
      style={style}
    >
      {/* Main Log Content */}
      <div className="p-2 flex items-start gap-2">
        <Button
          variant="ghost"
          size="sm"
          className="p-1 h-6 w-6"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          {isExpanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </Button>

        <div className="flex-1 space-y-1">
          {/* Primary Log Info */}
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-muted-foreground">
              {formatTime(log.time)}
            </span>
            <span
              className={cn("px-1.5 py-0.5 rounded-full text-xs font-medium", {
                "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-100":
                  log.levelname === "INFO",
                "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-100":
                  log.levelname === "ERROR",
                "bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-100":
                  log.levelname === "WARNING",
                "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-100":
                  log.levelname === "SUCCESS",
              })}
            >
              {log.levelname}
            </span>
            <span className="font-medium flex-1">{log.msg}</span>
          </div>

          {/* Secondary Info - Always Visible */}
          <div className="text-sm text-muted-foreground">
            <span className="font-mono">Process-{log.process}</span>
            <span className="mx-1">|</span>
            <span className="font-mono">{log.threadName}</span>
          </div>

          {/* Expanded Details */}
          {isExpanded && (
            <div className="mt-2 text-sm space-y-1 font-mono bg-background/50 rounded-md p-2">
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <span className="text-muted-foreground">Module: </span>
                  <span>{log.module}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Function: </span>
                  <span>{log.funcName}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Line: </span>
                  <span>{log.lineno}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Path: </span>
                  <span className="break-all">{log.pathname}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Process Name: </span>
                  <span>{log.processName}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Task: </span>
                  <span>{log.taskName || "None"}</span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
