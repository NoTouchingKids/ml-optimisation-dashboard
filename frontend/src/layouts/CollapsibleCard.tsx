import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import type { BaseProps } from "@/types/common";
import {
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from "@radix-ui/react-collapsible";
import { ChevronUp, ChevronDown } from "lucide-react";
import { useState } from "react";

interface CollapsibleCardProps extends BaseProps {
  title: string;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function CollapsibleCard({
  title,
  children,
  defaultOpen = true,
  onOpenChange,
}: CollapsibleCardProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open);
    onOpenChange?.(open);
  };

  return (
    <Collapsible open={isOpen} onOpenChange={handleOpenChange}>
      <Card>
        <CardHeader className="cursor-pointer">
          <div className="flex items-center justify-between">
            <CardTitle>{title}</CardTitle>
            <CollapsibleTrigger>
              {isOpen ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </CollapsibleTrigger>
          </div>
        </CardHeader>
        <CollapsibleContent>
          <CardContent>{children}</CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  );
}
