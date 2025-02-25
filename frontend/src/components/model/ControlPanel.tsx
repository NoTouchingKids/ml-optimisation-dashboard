import type { ModelStatus } from "@/types/common";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@radix-ui/react-select";
import { Square, Play } from "lucide-react";
import { Button } from "../ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "../ui/card";

interface ControlPanelProps {
  isRunning: boolean;
  onStartStop: () => void;
  status: ModelStatus;
  models: { value: string; label: string }[];
  selectedModel: string;
  onModelChange: (value: string) => void;
}

export function ControlPanel({
  isRunning,
  onStartStop,
  status,
  models,
  selectedModel,
  onModelChange,
}: ControlPanelProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Model Control</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-4 items-center">
          <Select value={selectedModel} onValueChange={onModelChange}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Select Model" />
            </SelectTrigger>
            <SelectContent>
              {models.map((model) => (
                <SelectItem key={model.value} value={model.value}>
                  {model.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Button
            onClick={onStartStop}
            variant={isRunning ? "destructive" : "default"}
            className="flex items-center gap-2"
          >
            {isRunning ? (
              <>
                <Square className="h-4 w-4" /> Stop
              </>
            ) : (
              <>
                <Play className="h-4 w-4" /> Start
              </>
            )}
          </Button>

          <div className="ml-4 text-sm">
            Status: <span className="font-semibold">{status}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
