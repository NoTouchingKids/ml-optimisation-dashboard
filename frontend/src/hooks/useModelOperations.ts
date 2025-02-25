import { useState } from "react";
import type { ModelAPI } from "../lib/model-api";
import type { ModelConfig, ModelStatus } from "../types/model";

export function useModelOperations(api: ModelAPI) {
  const [status, setStatus] = useState<ModelStatus | null>(null);
  const [error, setError] = useState<Error | null>(null);

  const trainModel = async (data: number[], config: ModelConfig) => {
    try {
      const result = await api.trainModel(data, config);
      setStatus(result);
      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      throw error;
    }
  };

  const predict = async (data: number[], config: ModelConfig) => {
    try {
      return await api.predict(data, config);
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      throw error;
    }
  };

  return { status, error, trainModel, predict };
}
