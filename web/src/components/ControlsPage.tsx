import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Gamepad2, Play, Square } from "lucide-react";

interface MachineControlState {
  machine_id: number;
  running: boolean;
  speed: number;
  setpoint: number;
  mode: string;
  temperature: number;
}

interface SystemResp {
  collector: { status: string };
}

const COMING_SOON = ["Reconnect", "Restart", "Shutdown"];

const MACHINE_NAMES: Record<number, string> = {
  1: "Fluid Bed Dryer",
  2: "Fluid Bed Processor",
  3: "Fluid Bed Equipment",
  4: "Tablet Coating Machine",
  5: "Compression Machine",
  6: "Compression Machine 2",
  7: "Tablet Printing Machine",
  8: "Capsule Checkweigher",
  9: "Rapid Mixer Granulator",
  10: "Blender 1",
  11: "Blender 2",
};

export default function ControlsPage() {
  const [controls, setControls] = useState<MachineControlState[]>([]);
  const [collectorStatus, setCollectorStatus] = useState("running");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchAll = useCallback(() => {
    setError("");
    Promise.all([
      fetch("/api/v1/controls").then((r) => { if (!r.ok) throw new Error("Controls fetch failed"); return r.json(); }),
      fetch("/system/status").then((r) => { if (!r.ok) throw new Error("System status fetch failed"); return r.json(); }),
    ])
      .then(([ctrl, sys]: [MachineControlState[], SystemResp]) => {
        setControls(ctrl ?? []);
        setCollectorStatus(sys.collector?.status ?? "unknown");
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchAll();
    const id = setInterval(fetchAll, 5000);
    return () => clearInterval(id);
  }, [fetchAll]);

  const sendAction = (id: number, action: string, value?: number) => {
    const url = value !== undefined
      ? `/api/v1/controls/${id}/${action}?value=${value}`
      : `/api/v1/controls/${id}/${action}`;
    fetch(url, { method: "POST" }).then(fetchAll);
  };

  const toggleCollector = () => {
    const url = collectorStatus === "running" ? "/collector/pause" : "/collector/resume";
    fetch(url, { method: "POST" }).then(fetchAll);
  };

  if (loading) {
    return (
      <main className="flex-1 h-full overflow-auto p-4 space-y-3">
        <div className="h-5 bg-secondary w-32 animate-pulse" />
        <div className="h-14 bg-secondary animate-pulse" />
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-32 bg-secondary animate-pulse" />
          ))}
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4 space-y-3">
      <h1 className="text-base font-semibold">Controls</h1>

      {error && (
        <div className="flex items-center gap-2 px-3 py-2 bg-red-900/30 border border-red-800 text-red-400 text-sm">
          <span>{error}</span>
          <Button size="sm" variant="outline" className="ml-auto" onClick={fetchAll}>Retry</Button>
        </div>
      )}

      <Card>
        <CardHeader className="py-2 px-3 min-h-0">
          <CardTitle>Data Collector</CardTitle>
        </CardHeader>
        <CardContent className="px-3 pb-3">
          <div className="flex items-center gap-3 flex-wrap">
            <Badge
              variant={collectorStatus === "running" ? "default" : "secondary"}
              className="text-[10px]"
            >
              {collectorStatus.toUpperCase()}
            </Badge>
            <Button
              size="sm"
              variant={collectorStatus === "running" ? "outline" : "default"}
              className="h-7 text-xs px-2"
              onClick={toggleCollector}
            >
              {collectorStatus === "running" ? "Pause" : "Resume"}
            </Button>
            <div className="flex-1" />
            {COMING_SOON.map((action) => (
              <Button key={action} size="sm" variant="outline" className="h-7 text-xs px-2 opacity-50" disabled title="Coming Soon">
                {action}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {!error && controls.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Gamepad2 size={32} className="text-muted-foreground/30 mb-3" />
          <p className="text-sm text-muted-foreground">No machine controls available.</p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2">
        {controls.map((c) => (
          <Card key={c.machine_id}>
            <CardHeader className="flex flex-row items-center justify-between py-2 px-3 min-h-0">
              <CardTitle className="text-sm truncate pr-2">{MACHINE_NAMES[c.machine_id] || `Machine ${c.machine_id}`}</CardTitle>
              <Badge variant={c.running ? "default" : "secondary"} className="text-[10px] shrink-0">
                {c.running ? "Running" : "Stopped"}
              </Badge>
            </CardHeader>
            <CardContent className="px-3 pb-3 space-y-2">
              <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs">
                <span className="text-muted-foreground">Speed</span>
                <span className="text-right font-mono text-foreground">{c.speed.toFixed(1)}</span>
                <span className="text-muted-foreground">Setpoint</span>
                <span className="text-right font-mono text-foreground">{c.setpoint.toFixed(1)}</span>
                <span className="text-muted-foreground">Mode</span>
                <span className="text-right font-mono text-foreground capitalize">{c.mode}</span>
                <span className="text-muted-foreground">Temperature</span>
                <span className="text-right font-mono text-foreground">{c.temperature.toFixed(1)}°C</span>
              </div>
              <div className="flex gap-1.5 pt-1">
                {c.running ? (
                  <Button
                    size="sm"
                    variant="destructive"
                    className="h-7 text-xs px-2"
                    onClick={() => sendAction(c.machine_id, "stop")}
                  >
                    <Square size={10} className="mr-1" /> Stop
                  </Button>
                ) : (
                  <Button
                    size="sm"
                    className="h-7 text-xs px-2"
                    onClick={() => sendAction(c.machine_id, "start")}
                  >
                    <Play size={10} className="mr-1" /> Start
                  </Button>
                )}
                <Button
                  size="sm"
                  variant="outline"
                  className="h-7 text-xs px-2"
                  onClick={() => {
                    const v = prompt("Enter setpoint:", String(c.setpoint));
                    if (v) sendAction(c.machine_id, "setpoint", Number(v));
                  }}
                >
                  Setpoint
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  className="h-7 text-xs px-2"
                  onClick={() => sendAction(c.machine_id, "mode")}
                >
                  Toggle Mode
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </main>
  );
}
