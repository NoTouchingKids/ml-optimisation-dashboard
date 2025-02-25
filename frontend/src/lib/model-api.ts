import type {
  ModelConfig,
  ModelRequest,
  ModelStatus,
  PredictionRequest,
  WebSocketHandler,
  WSMessage,
} from "@/types/model";
export interface ModelAPI {
  readonly clientId: string;
  readonly messageHandlers: Map<string, WebSocketHandler>;
  trainModel(data: number[], config: ModelConfig): Promise<ModelStatus>;
  predict(data: number[], config: ModelConfig): Promise<ModelStatus>;
  getModelStatus(): Promise<ModelStatus>;
  connectWebSocket<T>(
    messageType: string,
    handler: WebSocketHandler<T>
  ): () => void;
  disconnectWebSocket(): void;
  getClientId(): string;
}

export class ModelAPIImpl implements ModelAPI {
  public readonly clientId: string;
  public readonly messageHandlers: Map<string, WebSocketHandler>;
  private ws: WebSocket | null = null;
  private readonly baseUrl: string;

  constructor(baseUrl: string = "http://localhost:8080") {
    this.baseUrl = baseUrl;
    this.clientId = crypto.randomUUID();
    this.messageHandlers = new Map();
  }

  public trainModel(data: number[], config: ModelConfig): Promise<ModelStatus> {
    return this.makeRequest("/api/model/train", {
      type: "train",
      client_id: this.clientId,
      data,
      config,
    } as ModelRequest);
  }

  public predict(data: number[], config: ModelConfig): Promise<ModelStatus> {
    console.log({
      type: "predict",
      client_id: this.clientId,
      data,
      config,
    });
    return this.makeRequest("/api/model/predict", {
      type: "predict",
      client_id: this.clientId,
      data,
      config,
    } as ModelRequest);
  }

  public async getModelStatus(): Promise<ModelStatus> {
    const response = await fetch(
      `${this.baseUrl}/api/model/status/${this.clientId}`
    );
    if (!response.ok) {
      throw new Error(`Failed to get status: ${response.statusText}`);
    }
    return response.json();
  }

  public connectWebSocket<T>(
    messageType: string,
    handler: WebSocketHandler<T>
  ): () => void {
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      this.initializeWebSocket();
    }

    const wrappedHandler = (data: unknown) => handler(data as T);
    this.messageHandlers.set(messageType, wrappedHandler);
    return () => this.messageHandlers.delete(messageType);
  }

  public disconnectWebSocket(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  public getClientId(): string {
    return this.clientId;
  }

  private async makeRequest(
    endpoint: string,
    data: ModelRequest
  ): Promise<ModelStatus> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`Request failed: ${response.statusText}`);
    }

    return response.json();
  }

  private initializeWebSocket(): void {
    const wsUrl = `ws://${new URL(this.baseUrl).host}/ws?clientId=${
      this.clientId
    }`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data) as WSMessage;
      const handler = this.messageHandlers.get(message.type);
      if (handler) {
        handler(message.payload);
      }
    };

    this.ws.onclose = () => {
      setTimeout(() => this.initializeWebSocket(), 1000);
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
  }
}

export const modelApi = new ModelAPIImpl();
export default modelApi;
