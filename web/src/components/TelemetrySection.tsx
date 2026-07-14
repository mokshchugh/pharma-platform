import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import TelemetryChart from "./TelemetryChart";
import HistoricalTable from "./HistoricalTable";

interface AnalyticsPoint {
  timestamp: string;
  avg_value: number;
  min_value: number;
  max_value: number;
  sample_count: number;
}

interface TagAnalytics {
  tag_id: number;
  tag_name: string;
  data_type: string;
  unit: string;
  current: AnalyticsPoint | null;
  latest_value: number | null;
  series: AnalyticsPoint[];
  total_sample_count: number;
  window_avg: number;
  window_min: number;
  window_max: number;
}

export default function TelemetrySection({ tag }: { tag: TagAnalytics }) {
  const isAnalog = !["bool", "string", "bytes"].includes(tag.data_type);
  const [historyOpen, setHistoryOpen] = useState(false);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{tag.tag_name}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {isAnalog && tag.series.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-xs text-muted-foreground block">Average</span>
              <span className="font-semibold text-lg">
                {tag.window_avg.toFixed(2)} {tag.unit}
              </span>
            </div>
            <div>
              <span className="text-xs text-muted-foreground block">Minimum</span>
              <span className="font-semibold text-lg">
                {tag.window_min.toFixed(2)} {tag.unit}
              </span>
            </div>
            <div>
              <span className="text-xs text-muted-foreground block">Maximum</span>
              <span className="font-semibold text-lg">
                {tag.window_max.toFixed(2)} {tag.unit}
              </span>
            </div>
            <div>
              <span className="text-xs text-muted-foreground block">Samples</span>
              <span className="font-semibold text-lg">
                {tag.total_sample_count.toLocaleString()}
              </span>
            </div>
          </div>
        )}

        {isAnalog && tag.series.length === 0 && tag.current && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-xs text-muted-foreground block">Average</span>
              <span className="font-semibold text-lg">
                {tag.current.avg_value.toFixed(2)} {tag.unit}
              </span>
            </div>
            <div>
              <span className="text-xs text-muted-foreground block">Minimum</span>
              <span className="font-semibold text-lg">
                {tag.current.min_value.toFixed(2)} {tag.unit}
              </span>
            </div>
            <div>
              <span className="text-xs text-muted-foreground block">Maximum</span>
              <span className="font-semibold text-lg">
                {tag.current.max_value.toFixed(2)} {tag.unit}
              </span>
            </div>
            <div>
              <span className="text-xs text-muted-foreground block">Samples</span>
              <span className="font-semibold text-lg">
                {tag.current.sample_count.toLocaleString()}
              </span>
            </div>
          </div>
        )}

        {!isAnalog && tag.latest_value !== null && (
          <div className="text-sm">
            <span className="text-xs text-muted-foreground block">Latest State</span>
            <span className="font-semibold text-lg">
              {tag.latest_value === 1 ? "ON / TRUE" : "OFF / FALSE"}
            </span>
          </div>
        )}

        {isAnalog && tag.series.length > 0 && (
          <>
            <div>
              <h4 className="text-xs font-medium text-muted-foreground mb-2">Trend</h4>
              <TelemetryChart series={tag.series} unit={tag.unit} />
            </div>
            <div>
              <button
                onClick={() => setHistoryOpen((v) => !v)}
                className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors cursor-pointer border-none bg-transparent mb-2"
              >
                <span className="text-xs">{historyOpen ? "\u25BC" : "\u25B6"}</span>
                History
              </button>
              {historyOpen && <HistoricalTable series={tag.series} unit={tag.unit} />}
            </div>
          </>
        )}

        {isAnalog && tag.series.length === 0 && (
          <p className="text-sm text-muted-foreground">No historical data available for this time range.</p>
        )}
      </CardContent>
    </Card>
  );
}
