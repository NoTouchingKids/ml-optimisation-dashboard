export interface BaseProps {
  className?: string;
  children?: React.ReactNode;
}

export type ModelStatus =
  | "idle"
  | "training"
  | "predicting"
  | "error"
  | "complete";
