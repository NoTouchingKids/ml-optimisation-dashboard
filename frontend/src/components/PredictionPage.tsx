import React from "react";
import { PageContainer } from "@/layouts/PageContainer";
import { CollapsibleCard } from "@/layouts/CollapsibleCard";
import { ControlPanel } from "../components/model/ControlPanel";
import { VirtualizedLogs } from "./VirtualizedLogs";
import { ParameterForm } from "./ParameterForm";
import { useModelControl } from "../hooks/useModelControl";
import { useModelOperations } from "../hooks/useModelOperations";
import { useLogStream } from "@/hooks/useLogStream";
import { modelApi } from "../lib/model-api";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";

const AVAILABLE_MODELS = [
  { value: "basic", label: "Basic Model" },
  { value: "advanced", label: "Advanced Model" },
];

export default function PredictionPage() {
  const {
    isRunning,
    status,
    selectedModel,
    handleStartStop,
    setSelectedModel,
  } = useModelControl();

  const { predict, error } = useModelOperations(modelApi);
  const { logs } = useLogStream(modelApi, isRunning);

  const [paramsView, setParamsView] = React.useState<"ui" | "json">("ui");
  const [config, setConfig] = React.useState({
    featureEngineering: false,
    detrend: false,
    difference: false,
    windowSize: 10,
    threshold: 0.5,
  });

  const handlePredict = async () => {
    if (!selectedModel) return;
    handleStartStop();
    try {
      // Example data - replace with actual data source
      const data = Array.from({ length: 100 }, () => Math.random());
      await predict(data, config);
    } catch (err) {
      console.error("Prediction error:", err);
    }
  };

  return (
    <PageContainer>
      <ControlPanel
        isRunning={isRunning}
        onStartStop={handlePredict}
        status={status}
        models={AVAILABLE_MODELS}
        selectedModel={selectedModel}
        onModelChange={setSelectedModel}
      />

      <CollapsibleCard title="Model Parameters">
        <Tabs
          value={paramsView}
          onValueChange={(v) => setParamsView(v as "ui" | "json")}
        >
          <TabsList>
            <TabsTrigger value="ui">UI Form</TabsTrigger>
            <TabsTrigger value="json">JSON</TabsTrigger>
          </TabsList>
          <TabsContent value="ui">
            <ParameterForm config={config} onChange={setConfig} />
          </TabsContent>
          <TabsContent value="json">
            <ScrollArea className="h-64 w-full rounded-md border p-4">
              <pre className="text-sm">{JSON.stringify(config, null, 2)}</pre>
            </ScrollArea>
          </TabsContent>
        </Tabs>
      </CollapsibleCard>

      <CollapsibleCard title="Model Logs">
        <div className="h-96 border rounded-md">
          <VirtualizedLogs logs={logs || []} />
        </div>
      </CollapsibleCard>

      {error && (
        <Card>
          <CardContent className="text-destructive p-4">
            {error.message}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className="h-48 flex items-center justify-center text-muted-foreground">
          Prediction results will appear here
        </CardContent>
      </Card>
    </PageContainer>
  );
}
