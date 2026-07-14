import { useCallback, useEffect, useRef, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";

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
  const ms = diff;
  const sec = Math.floor(ms / 1000);
  const min = Math.floor(sec / 60);

  if (sec < 1) return "just now";
  if (sec < 5) return `${sec} sec ago`;
  if (sec < 60) return `${sec} sec ago`;
  if (min < 60) return `${min} min ago`;
  const hours = Math.floor(min / 60);
  if (hours < 24) return `${hours} hr ago`;
  const days = Math.floor(hours / 24);
  return `${days} day ago`;
}

export default function MachinesPage({ onNavigate }: { onNavigate?: (page: string, params?: Record<string, string>) => void }) {
  const [machines, setMachines] = useState<Machine[]>([]);
  const [error, setError] = useState("");
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef<ReturnType<typeof setInterval>>();

  const fetchMachines = useCallback(() => {
    setError("");
    fetch("/api/v1/machines")
      .then((r) => {
        if (!r.ok) throw new Error(`GET /api/v1/machines ${r.status}`);
        return r.json();
      })
      .then(setMachines)
      .catch((e) => setError(e.message));
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

  return (
    <main className="flex-1 h-full overflow-auto p-4">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-base font-semibold">Machines</h1>
        <button
          type="button"
          onClick={() => setAutoRefresh((v) => !v)}
          className={`text-xs px-2 py-1 rounded border ${
            autoRefresh ? "bg-green-100 border-green-300 text-green-700" : "bg-gray-100 border-gray-300 text-gray-500"
          }`}
        >
          Auto-refresh: {autoRefresh ? "ON" : "OFF"}
        </button>
      </div>

      {error && <p className="text-sm text-destructive mb-2">{error}</p>}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {machines.map((m) => (
          <button
            key={m.id}
            onClick={() => onNavigate?.("MachineDetail", { id: String(m.id) })}
            className="text-left cursor-pointer border-none bg-transparent p-0"
          >
            <Card className="transition-shadow hover:shadow-md h-full">
              <CardContent className="p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-semibold truncate">{m.machine_name}</p>
                  <span className="text-xs text-muted-foreground shrink-0">&rarr;</span>
                </div>

                <div className="flex items-center gap-3 text-xs">
                  <span className="flex items-center gap-1.5">
                    <span className={`w-2 h-2 rounded-full ${statusColor(m.status)}`} />
                    {m.status}
                  </span>
                  <span className="flex items-center gap-1.5">
                    <span className={`w-2 h-2 rounded-full ${collectionColor(m.collection_status)}`} />
                    {m.collection_status}
                  </span>
                </div>

                <div className="space-y-1 text-xs text-muted-foreground">
                  <div className="flex justify-between">
                    <span>PLC Make</span>
                    <span className="text-foreground">{m.plc_make}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>PLC Model</span>
                    <span className="text-foreground">{m.plc_model}</span>
                  </div>
                </div>

                <div className="text-xs text-muted-foreground">
                  Latest Sample: {m.last_sample ? relativeTime(m.last_sample) : "No telemetry"}
                </div>

                <div className="text-xs text-muted-foreground">
                  Tags: {m.enabled_tags} / {m.configured_tags}
                </div>
              </CardContent>
            </Card>
          </button>
        ))}
      </div>
    </main>
  );
}
