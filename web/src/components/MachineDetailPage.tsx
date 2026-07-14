import { useCallback, useEffect, useRef, useState } from "react";
import MachineSummaryCard from "./MachineSummaryCard";
import ResolutionSelector from "./ResolutionSelector";
import TimeWindowSelector from "./TimeWindowSelector";
import TelemetrySection from "./TelemetrySection";

interface Machine {
  id: number;
  machine_name: string;
  plc_make: string;
  plc_model: string;
  status: string;
  collection_status: string;
  last_sample: string | null;
  configured_tags: number;
  enabled_tags: number;
}

interface TelemetryTag {
  tag_id: number;
  tag_name: string;
  data_type: string;
  unit: string;
}

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

interface AnalyticsResponse {
  tags: TagAnalytics[];
}

function hoursAgo(n: number): string {
  const d = new Date(Date.now() - n * 3600_000);
  return d.toISOString();
}

function resolveWindow(preset: string, customFrom?: string, customTo?: string): { from: string; to: string } {
  switch (preset) {
    case "1h": return { from: hoursAgo(1), to: new Date().toISOString() };
    case "24h": return { from: hoursAgo(24), to: new Date().toISOString() };
    case "168h": return { from: hoursAgo(168), to: new Date().toISOString() };
    case "custom": return { from: customFrom ?? hoursAgo(1), to: customTo ?? new Date().toISOString() };
    default: return { from: hoursAgo(1), to: new Date().toISOString() };
  }
}

export default function MachineDetailPage({
  machineId,
  onBack,
}: {
  machineId: string;
  onBack?: () => void;
}) {
  const [machine, setMachine] = useState<Machine | null>(null);
  const [tags, setTags] = useState<TelemetryTag[]>([]);
  const [analytics, setAnalytics] = useState<AnalyticsResponse | null>(null);
  const [error, setError] = useState("");
  const [resolution, setResolution] = useState("1h");
  const [timeWindow, setTimeWindow] = useState("1h");
  const [customFrom, setCustomFrom] = useState<string>();
  const [customTo, setCustomTo] = useState<string>();
  const [loading, setLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<ReturnType<typeof setInterval>>();

  useEffect(() => {
    fetch(`/api/v1/machines/${machineId}`)
      .then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId} ${r.status}`);
        return r.json();
      })
      .then(setMachine)
      .catch((e) => setError(e.message));

    fetch(`/api/v1/machines/${machineId}/telemetry`)
      .then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId}/telemetry ${r.status}`);
        return r.json();
      })
      .then(setTags)
      .catch((e) => setError(e.message));
  }, [machineId]);

  const fetchAnalytics = useCallback(() => {
    setLoading(true);
    setError("");

    const { from, to } = resolveWindow(timeWindow, customFrom, customTo);
    const params = new URLSearchParams({ resolution, from, to });

    fetch(`/api/v1/machines/${machineId}/analytics?${params}`)
      .then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId}/analytics ${r.status}`);
        return r.json();
      })
      .then(setAnalytics)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [machineId, resolution, timeWindow, customFrom, customTo]);

  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  useEffect(() => {
    if (autoRefresh) {
      intervalRef.current = setInterval(fetchAnalytics, 5000);
    }
    return () => clearInterval(intervalRef.current);
  }, [autoRefresh, fetchAnalytics]);

  function handleTimeWindowChange(preset: string, from?: string, to?: string) {
    setTimeWindow(preset);
    if (preset === "custom") {
      setCustomFrom(from);
      setCustomTo(to);
    }
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4">
      <button
        onClick={onBack}
        className="text-xs text-muted-foreground hover:text-foreground transition-colors mb-3 cursor-pointer border-none bg-transparent"
      >
        &larr; Back to Machines
      </button>

      {error && <p className="text-sm text-destructive mb-2">{error}</p>}

      <div className="space-y-4">
        {machine && (
          <div className="max-w-xl">
            <h1 className="text-base font-semibold mb-3">{machine.machine_name}</h1>
            <MachineSummaryCard machine={machine} />
          </div>
        )}

        {tags.length > 0 && (
          <>
            <div className="flex items-center gap-4 flex-wrap">
              <ResolutionSelector value={resolution} onChange={setResolution} />
              <TimeWindowSelector value={timeWindow} onChange={handleTimeWindowChange} />
              <button
                type="button"
                onClick={() => setAutoRefresh((v) => !v)}
                className={`text-xs px-2 py-1 rounded border cursor-pointer ${
                  autoRefresh
                    ? "bg-green-100 border-green-300 text-green-700"
                    : "bg-gray-100 border-gray-300 text-gray-500"
                }`}
              >
                Auto-refresh: {autoRefresh ? "ON" : "OFF"}
              </button>
            </div>

            {analytics && analytics.tags.length > 0 && (
              <div className="space-y-6">
                {analytics.tags.map((tag) => (
                  <TelemetrySection key={tag.tag_id} tag={tag} />
                ))}
              </div>
            )}

            {analytics && analytics.tags.length === 0 && (
              <p className="text-sm text-muted-foreground">No telemetry tags found for this machine.</p>
            )}
          </>
        )}

        {!machine && !error && <p className="text-sm text-muted-foreground">Loading machine...</p>}
      </div>
    </main>
  );
}
