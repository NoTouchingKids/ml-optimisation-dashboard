import React from "react";
import { Button } from "@/components/ui/button";
import {
  BeakerIcon,
  ChartBarIcon,
  BrainIcon,
  LineChartIcon,
} from "lucide-react";

const Navigation = () => {
  return (
    <nav className="border-b">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          <div className="flex-shrink-0 flex items-center">
            <a href="/" className="flex items-center">
              <BeakerIcon className="h-8 w-8 text-primary" />
              <span className="ml-2 text-xl font-bold">ML Optimizer</span>
            </a>
          </div>

          <div className="hidden md:flex items-center space-x-4">
            <a href="/train">
              <Button variant="ghost" className="flex items-center">
                <BrainIcon className="h-5 w-5 mr-2" />
                Train
              </Button>
            </a>
            <a href="/predict">
              <Button variant="ghost" className="flex items-center">
                <LineChartIcon className="h-5 w-5 mr-2" />
                Predict
              </Button>
            </a>
            <a href="/results">
              <Button variant="ghost" className="flex items-center">
                <ChartBarIcon className="h-5 w-5 mr-2" />
                Results
              </Button>
            </a>
          </div>

          <div className="flex space-x-4">
            <Button
              variant="outline"
              onClick={() => (window.location.href = "/login")}
            >
              Sign Out
            </Button>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
