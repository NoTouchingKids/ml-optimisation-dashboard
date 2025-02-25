import { cn } from "@/lib/utils";
import type { BaseProps } from "@/types/common";

export function PageContainer({ children, className }: BaseProps) {
  return (
    <div className={cn("container mx-auto p-4 space-y-4", className)}>
      {children}
    </div>
  );
}
