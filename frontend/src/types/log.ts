export interface LogRecordPayload {
  timestamp: number;
  client_id: string;
  message: string;
  process_id: string;
}

export interface LogRecord {
  time: Date;
  status?: string;
  clientId: string;
  levelname: string;
  name?: string;
  levelno?: number;
  pathname?: string;
  filename?: string;
  module?: string;
  exc_info?: any;
  lineno?: number;
  msg: string;
  args?: any;
  exc_text?: string;
  funcName?: string;
  created?: number;
  msecs?: number;
  relativeCreated?: number;
  thread?: number;
  threadName?: string;
  process: number;
  processName: string;
  message?: string;
  [key: string]: any;
}
