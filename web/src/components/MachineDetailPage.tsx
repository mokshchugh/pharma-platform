import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "./ui/button";
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
  const [initialLoading, setInitialLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);

  useEffect(() => {
    let cancelled = false;

    Promise.all([
      fetch(`/api/v1/machines/${machineId}`).then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId} ${r.status}`);
        return r.json() as Promise<Machine>;
      }),
      fetch(`/api/v1/machines/${machineId}/telemetry`).then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId}/telemetry ${r.status}`);
        return r.json() as Promise<TelemetryTag[]>;
      }),
    ])
      .then(([m, t]) => {
        if (!cancelled) {
          setMachine(m);
          setTags(t);
        }
      })
      .catch((e) => {
        if (!cancelled) setError(e.message);
      })
      .finally(() => {
        if (!cancelled) setInitialLoading(false);
      });

    return () => { cancelled = true; };
  }, [machineId]);

  const retryInitial = useCallback(() => {
    setInitialLoading(true);
    setError("");

    Promise.all([
      fetch(`/api/v1/machines/${machineId}`).then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId} ${r.status}`);
        return r.json() as Promise<Machine>;
      }),
      fetch(`/api/v1/machines/${machineId}/telemetry`).then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId}/telemetry ${r.status}`);
        return r.json() as Promise<TelemetryTag[]>;
      }),
    ])
      .then(([m, t]) => {
        setMachine(m);
        setTags(t);
      })
      .catch((e) => setError(e.message))
      .finally(() => setInitialLoading(false));
  }, [machineId]);

  const fetchAnalytics = useCallback(() => {
    setLoading(true);
    setError("");

    const { from, to } = resolveWindow(timeWindow, customFrom, customTo);
    const params = new URLSearchParams({ resolution, from, to });

    fetch(`/api/v1/machines/${machineId}/analytics?${params}`)
      .then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines/${machineId}/analytics ${r.status}`);
        return r.json() as Promise<AnalyticsResponse>;
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

  if (initialLoading) {
    return (
      <main className="flex-1 h-full overflow-auto p-2">
        <div className="animate-pulse space-y-2">
          <div className="h-3 w-20 bg-muted rounded" />
          <div className="h-4 w-40 bg-muted rounded" />
          <div className="h-36 bg-muted rounded" />
          <div className="flex gap-2">
            <div className="h-6 w-36 bg-muted rounded" />
            <div className="h-6 w-52 bg-muted rounded" />
          </div>
          <div className="h-28 bg-muted rounded" />
          <div className="h-28 bg-muted rounded" />
        </div>
      </main>
    );
  }

  if (error && !machine) {
    return (
      <main className="flex-1 h-full overflow-auto p-2">
        <button
          onClick={onBack}
          className="text-xs text-muted-foreground hover:text-foreground transition-colors mb-1 cursor-pointer border-none bg-transparent"
        >
          &larr; Back to Machines
        </button>
        <div className="flex flex-col items-center justify-center py-10 gap-2">
          <p className="text-sm text-destructive">{error}</p>
          <Button variant="outline" size="sm" onClick={retryInitial}>Retry</Button>
        </div>
      </main>
    );
  }

  if (!machine && !error) {
    return (
      <main className="flex-1 h-full overflow-auto p-2">
        <button
          onClick={onBack}
          className="text-xs text-muted-foreground hover:text-foreground transition-colors mb-1 cursor-pointer border-none bg-transparent"
        >
          &larr; Back to Machines
        </button>
        <p className="text-xs text-muted-foreground">Machine not found.</p>
      </main>
    );
  }

  const m = machine!;
  return (
    <main className="flex-1 h-full overflow-auto p-2">
      <button
        onClick={onBack}
        className="text-xs text-muted-foreground hover:text-foreground transition-colors mb-0.5 cursor-pointer border-none bg-transparent"
      >
        &larr; Back to Machines
      </button>

      <h1 className="text-sm font-semibold mb-1">{m.machine_name}</h1>

      <div className="max-w-xl mb-1">
        <MachineSummaryCard machine={m} />
      </div>

      {tags.length > 0 && (
        <>
          <div className="flex items-center gap-1.5 flex-wrap mb-1">
            <ResolutionSelector value={resolution} onChange={setResolution} />
            <TimeWindowSelector value={timeWindow} onChange={handleTimeWindowChange} />
            <button
              type="button"
              onClick={() => setAutoRefresh((v) => !v)}
              className={`text-[11px] px-1.5 py-0.5 rounded border cursor-pointer leading-tight ${
                autoRefresh
                  ? "bg-green-100 border-green-300 text-green-700"
                  : "bg-gray-100 border-gray-300 text-gray-500"
              }`}
            >
              Auto: {autoRefresh ? "ON" : "OFF"}
            </button>
          </div>

          {error && (
            <div className="flex items-center gap-2 mb-1">
              <p className="text-[11px] text-destructive">{error}</p>
              <Button variant="outline" size="sm" onClick={fetchAnalytics}>Retry</Button>
            </div>
          )}

          {loading && !analytics && (
            <div className="animate-pulse space-y-1">
              <div className="h-24 bg-muted rounded" />
              <div className="h-24 bg-muted rounded" />
            </div>
          )}

          {analytics && analytics.tags.length > 0 && (
            <div className="space-y-0.5">
              {analytics.tags.map((tag) => (
                <TelemetrySection key={tag.tag_id} tag={tag} />
              ))}
            </div>
          )}

          {analytics && analytics.tags.length === 0 && (
            <p className="text-[11px] text-muted-foreground">No telemetry tags found for this machine.</p>
          )}
        </>
      )}

      {tags.length === 0 && !loading && (
        <p className="text-[11px] text-muted-foreground">No telemetry tags configured for this machine.</p>
      )}
    </main>
  );
}
