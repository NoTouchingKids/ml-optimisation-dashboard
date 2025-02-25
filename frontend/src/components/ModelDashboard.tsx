// src/components/ModelDashboard.tsx
import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import api, { type ModelStatus, type ModelConfig } from "../lib/api";
import { BeakerIcon, ArrowPathIcon } from "@heroicons/react/24/outline";

interface ModelConfigFormProps {
  config: ModelConfig;
  onChange: (config: ModelConfig) => void;
}

const ModelConfigForm: React.FC<ModelConfigFormProps> = ({
  config,
  onChange,
}) => {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">Feature Engineering</label>
        <Switch
          checked={config.featureEngineering}
          onCheckedChange={(checked) =>
            onChange({ ...config, featureEngineering: checked })
          }
        />
      </div>
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">Detrend</label>
        <Switch
          checked={config.detrend}
          onCheckedChange={(checked) =>
            onChange({ ...config, detrend: checked })
          }
        />
      </div>
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">Difference</label>
        <Switch
          checked={config.difference}
          onCheckedChange={(checked) =>
            onChange({ ...config, difference: checked })
          }
        />
      </div>
    </div>
  );
};

const ModelDashboard: React.FC = () => {
  const [status, setStatus] = useState<ModelStatus | null>(null);
  const [config, setConfig] = useState<ModelConfig>({
    featureEngineering: true,
    detrend: false,
    difference: false,
  });
  const [isTraining, setIsTraining] = useState(false);

  useEffect(() => {
    // Connect to WebSocket for real-time updates
    api.connectWebSocket((data) => {
      if (data.type === "model_status") {
        setStatus(data.payload);
      }
    });

    // Cleanup on unmount
    return () => {
      api.disconnectWebSocket();
    };
  }, []);

  const handleTrain = async () => {
    setIsTraining(true);
    try {
      // Mock data for example
      const data = Array.from({ length: 100 }, () => Math.random());
      await api.trainModel(data, config);
    } catch (error) {
      console.error("Training error:", error);
    } finally {
      setIsTraining(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <Tabs defaultValue="overview" className="space-y-8">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="results">Results</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <BeakerIcon className="h-5 w-5 mr-2" />
                  Model Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">
                    Current Status: {status?.status || "Not Started"}
                  </p>
                  {status?.message && (
                    <p className="text-sm text-muted-foreground">
                      {status.message}
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <Button
                    onClick={handleTrain}
                    disabled={isTraining}
                    className="w-full"
                  >
                    {isTraining ? (
                      <>
                        <ArrowPathIcon className="h-4 w-4 mr-2 animate-spin" />
                        Training...
                      </>
                    ) : (
                      "Start Training"
                    )}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="configuration">
          <Card>
            <CardHeader>
              <CardTitle>Model Configuration</CardTitle>
            </CardHeader>
            <CardContent>
              <ModelConfigForm config={config} onChange={setConfig} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="results">
          <Card>
            <CardHeader>
              <CardTitle>Training Results</CardTitle>
            </CardHeader>
            <CardContent>
              {/* Add visualization components here */}
              <p className="text-muted-foreground">
                Visualization components will be added here
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default ModelDashboard;
