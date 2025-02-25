import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { v4 as uuidv4 } from "uuid";
import { Check, Copy, RefreshCw } from "lucide-react";

interface ClientIdInputProps {
  onIdChange: (id: string) => void;
  disabled?: boolean;
}

export function ClientIdInput({ onIdChange, disabled }: ClientIdInputProps) {
  const [clientId, setClientId] = useState(uuidv4());
  const [copied, setCopied] = useState(false);

  const generateNewId = () => {
    const newId = uuidv4();
    setClientId(newId);
    onIdChange(newId);
  };

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(clientId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex items-center space-x-2">
      <Input
        value={clientId}
        readOnly
        className="font-mono"
        disabled={disabled}
      />
      <Button
        size="icon"
        variant="outline"
        onClick={copyToClipboard}
        disabled={disabled}
      >
        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
      </Button>
      <Button
        size="icon"
        variant="outline"
        onClick={generateNewId}
        disabled={disabled}
      >
        <RefreshCw className="h-4 w-4" />
      </Button>
    </div>
  );
}
