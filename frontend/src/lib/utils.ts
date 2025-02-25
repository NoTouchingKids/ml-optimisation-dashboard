import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import * as fzstd from "fzstd";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function base64ToUint8Array(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}

export function decompressData(compressedData: Uint8Array): Uint8Array {
  try {
    return fzstd.decompress(compressedData);
  } catch (error) {
    console.error("Decompression error:", error);
    throw error;
  }
}
