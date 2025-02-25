import type { ModelStatus } from "@/types/common";
import { useState } from "react";

interface ModelControlState {
  isRunning: boolean;
  status: ModelStatus;
  selectedModel: string;
}

interface ModelControlOptions {
  onStatusChange?: (status: ModelStatus) => void;
}

export function useModelControl(options?: ModelControlOptions) {
  const [state, setState] = useState<ModelControlState>({
    isRunning: false,
    status: "idle",
    selectedModel: "",
  });

  const handleStartStop = () => {
    setState((prev) => ({
      ...prev,
      isRunning: !prev.isRunning,
      status: !prev.isRunning ? "training" : "idle",
    }));
    options?.onStatusChange?.(state.isRunning ? "idle" : "training");
  };

  const setSelectedModel = (model: string) => {
    setState((prev) => ({ ...prev, selectedModel: model }));
  };

  return {
    ...state,
    handleStartStop,
    setSelectedModel,
  };
}
