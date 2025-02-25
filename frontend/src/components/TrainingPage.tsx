import { VirtualizedLogs } from "./VirtualizedLogs";
import { HyperParameterForm } from "./HyperParameterForm";
import { useModelControl } from "@/hooks/useModelControl";
import { PageContainer } from "@/layouts/PageContainer";
import { ControlPanel } from "./model/ControlPanel";
import { CollapsibleCard } from "@/layouts/CollapsibleCard";
import { modelApi } from "@/lib/model-api";
import { useLogStream } from "@/hooks/useLogStream";

export default function TrainingPage() {
  const {
    isRunning,
    status,
    selectedModel,
    handleStartStop,
    setSelectedModel,
  } = useModelControl();
  const { logs } = useLogStream(modelApi, isRunning);

  return (
    <PageContainer>
      <ControlPanel
        isRunning={isRunning}
        onStartStop={handleStartStop}
        status={status}
        models={[{ value: "model1", label: "Basic Model" }]}
        selectedModel={selectedModel}
        onModelChange={setSelectedModel}
      />

      <CollapsibleCard title="Model Parameters">
        <HyperParameterForm />
      </CollapsibleCard>

      <CollapsibleCard title="Training Logs">
        <div className="h-auto border rounded-md">
          <VirtualizedLogs logs={logs} />
        </div>
      </CollapsibleCard>
    </PageContainer>
  );
}
