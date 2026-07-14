import { useCallback, useEffect, useRef, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Cpu } from "lucide-react";

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

function statusColor(status: string): string {
  switch (status) {
    case "ONLINE": return "bg-green-500";
    case "OFFLINE": return "bg-orange-500";
    case "ERROR": return "bg-red-500";
    default: return "bg-gray-400";
  }
}

function collectionColor(status: string): string {
  switch (status) {
    case "COLLECTING": return "bg-green-500";
    case "PAUSED": return "bg-yellow-500";
    default: return "bg-gray-400";
  }
}

function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime();
  const sec = Math.floor(diff / 1000);
  const min = Math.floor(sec / 60);

  if (sec < 1) return "just now";
  if (sec < 60) return `${sec}s ago`;
  if (min < 60) return `${min}m ago`;
  const hours = Math.floor(min / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function SkeletonCard() {
  return (
    <div className="border border-border bg-card p-4 space-y-3 animate-pulse">
      <div className="h-4 bg-secondary rounded w-3/4" />
      <div className="flex gap-3">
        <div className="h-3 bg-secondary rounded w-16" />
        <div className="h-3 bg-secondary rounded w-20" />
      </div>
      <div className="space-y-1">
        <div className="h-3 bg-secondary rounded w-full" />
        <div className="h-3 bg-secondary rounded w-2/3" />
      </div>
      <div className="h-3 bg-secondary rounded w-1/2" />
      <div className="h-3 bg-secondary rounded w-1/3" />
    </div>
  );
}

export default function MachinesPage({ onNavigate }: { onNavigate?: (page: string, params?: Record<string, string>) => void }) {
  const [machines, setMachines] = useState<Machine[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);

  const fetchMachines = useCallback(() => {
    setError("");
    fetch("/api/v1/machines")
      .then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines ${r.status}`);
        return r.json();
      })
      .then((data) => {
        setMachines(data);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchMachines();
  }, [fetchMachines]);

  useEffect(() => {
    if (autoRefresh) {
      intervalRef.current = setInterval(fetchMachines, 5000);
    }
    return () => clearInterval(intervalRef.current);
  }, [autoRefresh, fetchMachines]);

  if (loading) {
    return (
      <main className="flex-1 h-full overflow-auto p-4">
        <div className="flex items-center justify-between mb-4">
          <div className="h-5 bg-secondary rounded w-24 animate-pulse" />
          <div className="h-6 bg-secondary rounded w-28 animate-pulse" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          {Array.from({ length: 8 }).map((_, i) => <SkeletonCard key={i} />)}
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <h1 className="text-base font-semibold">Machines</h1>
          <span className="text-xs text-muted-foreground">{machines.length} configured</span>
        </div>
        <button
          type="button"
          onClick={() => setAutoRefresh((v) => !v)}
          className={`text-xs px-2 py-1 border cursor-pointer transition-colors ${
            autoRefresh ? "bg-green-50 border-green-300 text-green-700" : "bg-gray-50 border-gray-300 text-gray-500"
          }`}
        >
          Auto: {autoRefresh ? "ON" : "OFF"}
        </button>
      </div>

      {error && (
        <div className="flex items-center gap-2 px-3 py-2 mb-3 bg-red-50 border border-red-200 text-red-700 text-sm">
          <span>{error}</span>
          <button
            onClick={fetchMachines}
            className="ml-auto text-xs px-2 py-0.5 border border-red-300 hover:bg-red-100 cursor-pointer bg-transparent"
          >
            Retry
          </button>
        </div>
      )}

      {!error && machines.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Cpu size={32} className="text-muted-foreground/40 mb-3" />
          <p className="text-sm text-muted-foreground">No machines configured.</p>
          <p className="text-xs text-muted-foreground mt-1">Machines will appear here once they are added to the system.</p>
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
        {machines.map((m) => (
          <button
            key={m.id}
            onClick={() => onNavigate?.("MachineDetail", { id: String(m.id) })}
            className="text-left cursor-pointer border-none bg-transparent p-0"
          >
            <Card className="transition-shadow hover:shadow-md h-full">
              <CardContent className="p-3 space-y-2.5">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-semibold truncate">{m.machine_name}</p>
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-muted-foreground shrink-0">
                    <path d="M9 18l6-6-6-6" />
                  </svg>
                </div>

                <div className="flex items-center gap-3 text-[11px]">
                  <span className="flex items-center gap-1.5">
                    <span className={`w-1.5 h-1.5 rounded-full ${statusColor(m.status)}`} />
                    {m.status}
                  </span>
                  <span className="flex items-center gap-1.5">
                    <span className={`w-1.5 h-1.5 rounded-full ${collectionColor(m.collection_status)}`} />
                    {m.collection_status}
                  </span>
                </div>

                <div className="space-y-1 text-[11px] text-muted-foreground">
                  <div className="flex justify-between">
                    <span>PLC Make</span>
                    <span className="text-foreground">{m.plc_make || "—"}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>PLC Model</span>
                    <span className="text-foreground">{m.plc_model || "—"}</span>
                  </div>
                </div>

                <div className="text-[11px] text-muted-foreground">
                  Latest: {m.last_sample ? relativeTime(m.last_sample) : "No telemetry"}
                </div>

                <div className="text-[11px] text-muted-foreground">
                  Tags: {m.enabled_tags}/{m.configured_tags} enabled
                </div>
              </CardContent>
            </Card>
          </button>
        ))}
      </div>
    </main>
  );
}
