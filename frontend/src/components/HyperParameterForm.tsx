import React from "react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Slider } from "@/components/ui/slider";
import { Card, CardContent } from "@/components/ui/card";
import * as z from "zod";

const hyperParamSchema = z.object({
  learningRate: z.number().min(0.0001).max(1),
  batchSize: z.number().min(1).max(512),
  epochs: z.number().min(1).max(1000),
  dropoutRate: z.number().min(0).max(1),
  layerSizes: z.array(z.number().min(1).max(1024)),
});

type HyperParameters = z.infer<typeof hyperParamSchema>;

interface HyperParameterFormProps {
  onChange?: (params: HyperParameters) => void;
}

export function HyperParameterForm({ onChange }: HyperParameterFormProps) {
  const [params, setParams] = React.useState<HyperParameters>({
    learningRate: 0.001,
    batchSize: 32,
    epochs: 100,
    dropoutRate: 0.2,
    layerSizes: [64, 32],
  });

  const handleChange = (key: keyof HyperParameters, value: any) => {
    const newParams = { ...params, [key]: value };
    try {
      hyperParamSchema.parse(newParams);
      setParams(newParams);
      onChange?.(newParams);
    } catch (error) {
      console.error("Validation error:", error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <Label>Learning Rate</Label>
        <Slider
          value={[params.learningRate]}
          min={0.0001}
          max={1}
          step={0.0001}
          onValueChange={([value]) => handleChange("learningRate", value)}
        />
        <div className="text-sm text-muted-foreground">
          Current: {params.learningRate}
        </div>
      </div>

      <div className="space-y-4">
        <Label>Batch Size</Label>
        <Input
          type="number"
          value={params.batchSize}
          onChange={(e) => handleChange("batchSize", parseInt(e.target.value))}
          min={1}
          max={512}
        />
      </div>

      <div className="space-y-4">
        <Label>Epochs</Label>
        <Input
          type="number"
          value={params.epochs}
          onChange={(e) => handleChange("epochs", parseInt(e.target.value))}
          min={1}
          max={1000}
        />
      </div>

      <div className="space-y-4">
        <Label>Dropout Rate</Label>
        <Slider
          value={[params.dropoutRate]}
          min={0}
          max={1}
          step={0.01}
          onValueChange={([value]) => handleChange("dropoutRate", value)}
        />
        <div className="text-sm text-muted-foreground">
          Current: {params.dropoutRate}
        </div>
      </div>

      <div className="space-y-4">
        <Label>Layer Sizes</Label>
        {params.layerSizes.map((size, index) => (
          <Input
            key={index}
            type="number"
            value={size}
            onChange={(e) => {
              const newSizes = [...params.layerSizes];
              newSizes[index] = parseInt(e.target.value);
              handleChange("layerSizes", newSizes);
            }}
            min={1}
            max={1024}
          />
        ))}
      </div>
    </div>
  );
}
