import { Card, CardContent } from "@/components/ui/card";
import type { LogMetrics } from "@/types/process";

interface MetricsPanelProps {
  metrics: LogMetrics;
}

export function MetricsPanel({ metrics }: MetricsPanelProps) {
  const runtime = metrics.startTime
    ? Math.round((Date.now() - metrics.startTime.getTime()) / 1000)
    : 0;

  const errorRate = metrics.totalLogs
    ? ((metrics.errorCount / metrics.totalLogs) * 100).toFixed(1)
    : "0";

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card>
        <CardContent className="pt-6">
          <div className="text-sm font-medium text-muted-foreground">
            Total Logs
          </div>
          <div className="text-2xl font-bold mt-2">{metrics.totalLogs}</div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="pt-6">
          <div className="text-sm font-medium text-muted-foreground">
            Error Rate
          </div>
          <div className="text-2xl font-bold mt-2">{errorRate}%</div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="pt-6">
          <div className="text-sm font-medium text-muted-foreground">
            Status
          </div>
          <div className="text-2xl font-bold mt-2 capitalize">
            {metrics.status}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="pt-6">
          <div className="text-sm font-medium text-muted-foreground">
            Runtime
          </div>
          <div className="text-2xl font-bold mt-2">{runtime}s</div>
        </CardContent>
      </Card>
    </div>
  );
}
