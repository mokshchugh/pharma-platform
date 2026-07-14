import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Factory } from "lucide-react";

interface ProductionRun {
  id: number;
  machine_id: number;
  machine_name: string;
  batch_id: string;
  product_name: string;
  target_qty: number;
  good_qty: number;
  bad_qty: number;
  start_time: string;
  end_time?: string;
  status: string;
  created_at: string;
}

interface DowntimeEvent {
  id: number;
  machine_id: number;
  machine_name: string;
  start_time: string;
  end_time?: string;
  reason: string;
  category: string;
  duration_seconds: number;
  created_at: string;
}

export default function ProductionPage() {
  const [runs, setRuns] = useState<ProductionRun[]>([]);
  const [downtime, setDowntime] = useState<DowntimeEvent[]>([]);
  const [tab, setTab] = useState<"runs" | "downtime">("runs");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchAll = useCallback(() => {
    setError("");
    Promise.all([
      fetch("/api/v1/production").then((r) => { if (!r.ok) throw new Error("Production fetch failed"); return r.json(); }),
      fetch("/api/v1/downtime").then((r) => { if (!r.ok) throw new Error("Downtime fetch failed"); return r.json(); }),
    ])
      .then(([runsData, downtimeData]) => {
        setRuns(Array.isArray(runsData) ? runsData : []);
        setDowntime(Array.isArray(downtimeData) ? downtimeData : []);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchAll();
    const id = setInterval(fetchAll, 10000);
    return () => clearInterval(id);
  }, [fetchAll]);

  if (loading) {
    return (
      <main className="flex-1 h-full overflow-auto p-4 space-y-3">
        <div className="h-5 bg-secondary w-24 animate-pulse" />
        <div className="flex gap-2">
          <div className="h-7 bg-secondary w-32 animate-pulse" />
          <div className="h-7 bg-secondary w-24 animate-pulse" />
        </div>
        <div className="border border-border">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-10 bg-secondary/30 border-b border-border animate-pulse" />
          ))}
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4 space-y-3">
      <h1 className="text-base font-semibold">Production</h1>

      {error && (
        <div className="flex items-center gap-2 px-3 py-2 bg-red-900/30 border border-red-800 text-red-400 text-sm">
          <span>{error}</span>
          <Button size="sm" variant="outline" className="ml-auto" onClick={fetchAll}>Retry</Button>
        </div>
      )}

      <div className="flex gap-2">
        <div className="flex border border-input rounded overflow-hidden">
          <button
            onClick={() => setTab("runs")}
            className={`px-3 py-1 text-xs font-medium border-r last:border-r-0 cursor-pointer transition-colors ${
              tab === "runs" ? "bg-primary text-primary-foreground" : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            Runs {runs.length > 0 && <span className="ml-1 opacity-60">({runs.length})</span>}
          </button>
          <button
            onClick={() => setTab("downtime")}
            className={`px-3 py-1 text-xs font-medium border-r last:border-r-0 cursor-pointer transition-colors ${
              tab === "downtime" ? "bg-primary text-primary-foreground" : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            Downtime {downtime.length > 0 && <span className="ml-1 opacity-60">({downtime.length})</span>}
          </button>
        </div>
      </div>

      {tab === "runs" && (
        <Card>
          <CardHeader className="py-2 px-3 min-h-0"><CardTitle>Production Runs</CardTitle></CardHeader>
          <CardContent className="p-0">
            {!error && runs.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <Factory size={28} className="text-muted-foreground/30 mb-2" />
                <p className="text-sm text-muted-foreground">No production runs found.</p>
                <p className="text-xs text-muted-foreground/60 mt-1">Active runs will appear here once production starts.</p>
              </div>
            ) : (
              <div className="overflow-auto max-h-[65vh]">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border bg-muted sticky top-0">
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Machine</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Product</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Batch</th>
                      <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Good</th>
                      <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Bad</th>
                      <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Target</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Progress</th>
                      <th className="text-center px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {runs.map((r) => {
                      const total = r.good_qty + r.bad_qty;
                      const pct = r.target_qty > 0 ? Math.min(100, (total / r.target_qty) * 100) : 0;
                      return (
                        <tr key={r.id} className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors">
                          <td className="px-3 py-1.5 text-sm">{r.machine_name}</td>
                          <td className="px-3 py-1.5 text-sm">{r.product_name}</td>
                          <td className="px-3 py-1.5 font-mono text-xs">{r.batch_id}</td>
                          <td className="px-3 py-1.5 text-right text-green-500 font-mono text-xs">{r.good_qty}</td>
                          <td className="px-3 py-1.5 text-right text-red-500 font-mono text-xs">{r.bad_qty}</td>
                          <td className="px-3 py-1.5 text-right font-mono text-xs text-muted-foreground">{r.target_qty}</td>
                          <td className="px-3 py-1.5">
                            <div className="flex items-center gap-2">
                              <div className="w-16 h-1.5 bg-secondary overflow-hidden">
                                <div className="h-full bg-primary" style={{ width: `${pct}%` }} />
                              </div>
                              <span className="text-[10px] text-muted-foreground font-mono">{pct.toFixed(0)}%</span>
                            </div>
                          </td>
                          <td className="px-3 py-1.5 text-center">
                            <Badge variant={r.status === "running" ? "default" : "secondary"} className="text-[10px]">{r.status}</Badge>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {tab === "downtime" && (
        <Card>
          <CardHeader className="py-2 px-3 min-h-0"><CardTitle>Downtime Events</CardTitle></CardHeader>
          <CardContent className="p-0">
            {!error && downtime.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <Factory size={28} className="text-muted-foreground/30 mb-2" />
                <p className="text-sm text-muted-foreground">No downtime events recorded.</p>
              </div>
            ) : (
              <div className="overflow-auto max-h-[65vh]">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border bg-muted sticky top-0">
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Machine</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Category</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Reason</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Start</th>
                      <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">End</th>
                      <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Duration</th>
                    </tr>
                  </thead>
                  <tbody>
                    {downtime.map((d) => (
                      <tr key={d.id} className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors">
                        <td className="px-3 py-1.5 text-sm">{d.machine_name}</td>
                        <td className="px-3 py-1.5">
                          <Badge variant={d.category === "breakdown" ? "destructive" : d.category === "changeover" ? "secondary" : "outline"} className="text-[10px]">
                            {d.category}
                          </Badge>
                        </td>
                        <td className="px-3 py-1.5 text-sm">{d.reason}</td>
                        <td className="px-3 py-1.5 text-xs whitespace-nowrap text-muted-foreground">
                          {new Date(d.start_time).toLocaleString()}
                        </td>
                        <td className="px-3 py-1.5 text-xs whitespace-nowrap text-muted-foreground">
                          {d.end_time ? new Date(d.end_time).toLocaleString() : "—"}
                        </td>
                        <td className="px-3 py-1.5 text-right font-mono text-xs">
                          {d.duration_seconds > 0 ? `${Math.round(d.duration_seconds / 60)}m ${d.duration_seconds % 60}s` : "—"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </main>
  );
}
