import { base64ToUint8Array, decompressData } from "@/lib/utils";
import type { LogRecord } from "@/types/log";
import * as msgpack from "msgpack-lite";

const UUID_LENGTH = 36; // Standard UUID string length

export const processLog = async (raw: any): Promise<LogRecord | null> => {
  try {
    const logRecord = JSON.parse(raw).payload;
    const compressedData = await base64ToUint8Array(logRecord.message);
    const decompressedData = await decompressData(
      compressedData.slice(UUID_LENGTH)
    );
    return msgpack.decode(decompressedData) as LogRecord;
  } catch (error) {
    console.error("Error processing log:", error);
    return null;
  }
};
