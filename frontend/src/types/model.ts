import type { LogRecordPayload } from "./log";

export interface ModelConfig {
  featureEngineering: boolean;
  detrend: boolean;
  difference: boolean;
  windowSize?: number;
  threshold?: number;
}

export interface ModelRequest {
  type: "train" | "predict";
  client_id: string;
  data?: number[];
  startDate?: string;
  endDate?: string;
  config?: ModelConfig;
}

export interface ModelStatus {
  clientId: string;
  status: "idle" | "training" | "predicting" | "error" | "complete";
  message?: string;
  timestamp: string;
  processType: string;
}

export interface PredictionRequest {
  type: string;
  clientId: string;
  data: number[];
  config: ModelConfig;
}

// types/websocket.ts
export interface WSMessage {
  type: "live_log" | "model_status" | "history_request";
  payload: LogRecordPayload;
  requestId?: string;
}

export interface WebSocketHandler<T = unknown> {
  (data: T): void;
}
