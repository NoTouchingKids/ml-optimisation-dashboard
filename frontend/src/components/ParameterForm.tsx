import React from "react";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Slider } from "@/components/ui/slider";
import * as z from "zod";

const configSchema = z.object({
  featureEngineering: z.boolean(),
  detrend: z.boolean(),
  difference: z.boolean(),
  windowSize: z.number().min(1).max(100),
  threshold: z.number().min(0).max(1),
});

type ModelConfig = z.infer<typeof configSchema>;

interface ParameterFormProps {
  config: ModelConfig;
  onChange: (config: ModelConfig) => void;
}

export function ParameterForm({ config, onChange }: ParameterFormProps) {
  const handleChange = (key: keyof ModelConfig, value: any) => {
    const newConfig = { ...config, [key]: value };
    try {
      configSchema.parse(newConfig);
      onChange(newConfig);
    } catch (error) {
      console.error("Validation error:", error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Label htmlFor="featureEngineering">Feature Engineering</Label>
          <Switch
            id="featureEngineering"
            checked={config.featureEngineering}
            onCheckedChange={(checked) =>
              handleChange("featureEngineering", checked)
            }
          />
        </div>

        <div className="flex items-center justify-between">
          <Label htmlFor="detrend">Detrend Data</Label>
          <Switch
            id="detrend"
            checked={config.detrend}
            onCheckedChange={(checked) => handleChange("detrend", checked)}
          />
        </div>

        <div className="flex items-center justify-between">
          <Label htmlFor="difference">Use Difference</Label>
          <Switch
            id="difference"
            checked={config.difference}
            onCheckedChange={(checked) => handleChange("difference", checked)}
          />
        </div>
      </div>

      <div className="space-y-4">
        <Label>Window Size</Label>
        <Slider
          value={[config.windowSize]}
          min={1}
          max={100}
          step={1}
          onValueChange={([value]) => handleChange("windowSize", value)}
        />
        <div className="text-sm text-muted-foreground">
          Current: {config.windowSize}
        </div>
      </div>

      <div className="space-y-4">
        <Label>Threshold</Label>
        <Slider
          value={[config.threshold]}
          min={0}
          max={1}
          step={0.01}
          onValueChange={([value]) => handleChange("threshold", value)}
        />
        <div className="text-sm text-muted-foreground">
          Current: {config.threshold.toFixed(2)}
        </div>
      </div>
    </div>
  );
}
