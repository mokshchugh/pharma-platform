import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { AlertTriangle, Search } from "lucide-react";

interface Alarm {
  id: string;
  plc_id: string;
  tag_id?: string;
  message: string;
  severity: string;
  active: boolean;
  created_at: string;
  acknowledged_at?: string;
}

const SEVERITIES = ["all", "critical", "warning", "info"] as const;

function sevVariant(s: string) {
  switch (s) {
    case "critical": return "destructive" as const;
    case "warning": return "secondary" as const;
    default: return "outline" as const;
  }
}

export default function AlarmsPage() {
  const [alarms, setAlarms] = useState<Alarm[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showActive, setShowActive] = useState(true);
  const [severityFilter, setSeverityFilter] = useState("all");
  const [search, setSearch] = useState("");

  const fetchAlarms = useCallback(() => {
    setError("");
    const url = showActive ? "/alarms/active" : "/alarms";
    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`GET ${url} ${r.status}`);
        return r.json();
      })
      .then((data) => {
        setAlarms(data ?? []);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  }, [showActive]);

  useEffect(() => {
    setLoading(true);
    fetchAlarms();
    const id = setInterval(fetchAlarms, 5000);
    return () => clearInterval(id);
  }, [fetchAlarms]);

  const acknowledge = (id: string) => {
    fetch(`/alarms/acknowledge/${id}`, { method: "POST" }).then(fetchAlarms);
  };

  const filtered = alarms
    .filter((a) => severityFilter === "all" || a.severity === severityFilter)
    .filter((a) => !search || a.message.toLowerCase().includes(search.toLowerCase()) || a.plc_id.toLowerCase().includes(search.toLowerCase()));

  const counts = {
    critical: alarms.filter((a) => a.severity === "critical").length,
    warning: alarms.filter((a) => a.severity === "warning").length,
    info: alarms.filter((a) => a.severity === "info").length,
  };

  if (loading) {
    return (
      <main className="flex-1 h-full overflow-auto p-4 space-y-3">
        <div className="h-5 bg-secondary w-16 animate-pulse" />
        <div className="flex gap-2">
          <div className="h-7 bg-secondary w-16 animate-pulse" />
          <div className="h-7 bg-secondary w-16 animate-pulse" />
        </div>
        <div className="flex gap-2">
          <div className="h-7 bg-secondary w-20 animate-pulse" />
          <div className="h-7 bg-secondary w-24 animate-pulse" />
          <div className="h-7 bg-secondary w-16 animate-pulse" />
          <div className="h-7 bg-secondary w-48 animate-pulse ml-auto" />
        </div>
        <div className="border border-border">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-10 bg-secondary/30 border-b border-border animate-pulse" />
          ))}
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4 space-y-3">
      <h1 className="text-base font-semibold">Alarms</h1>

      {error && (
        <div className="flex items-center gap-2 px-3 py-2 bg-red-900/30 border border-red-800 text-red-400 text-sm">
          <span>{error}</span>
          <Button size="sm" variant="outline" className="ml-auto" onClick={fetchAlarms}>Retry</Button>
        </div>
      )}

      <div className="flex items-center gap-2 flex-wrap">
        <div className="flex border border-input rounded overflow-hidden">
          <button
            onClick={() => setShowActive(true)}
            className={`px-3 py-1 text-xs font-medium border-r last:border-r-0 cursor-pointer transition-colors ${
              showActive ? "bg-primary text-primary-foreground" : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            Active {alarms.filter((a) => a.active).length > 0 && `(${alarms.filter((a) => a.active).length})`}
          </button>
          <button
            onClick={() => setShowActive(false)}
            className={`px-3 py-1 text-xs font-medium border-r last:border-r-0 cursor-pointer transition-colors ${
              !showActive ? "bg-primary text-primary-foreground" : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            History
          </button>
        </div>
        <div className="w-px h-5 bg-border" />
        {SEVERITIES.map((s) => (
          <button
            key={s}
            onClick={() => setSeverityFilter(s)}
            className={`px-2.5 py-1 text-xs font-medium border border-input cursor-pointer transition-colors ${
              severityFilter === s ? "bg-primary text-primary-foreground border-primary" : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            {s === "all" ? "All" : s.charAt(0).toUpperCase() + s.slice(1)}
            {s !== "all" && counts[s as keyof typeof counts] > 0 && (
              <span className="ml-1 opacity-60">{counts[s as keyof typeof counts]}</span>
            )}
          </button>
        ))}
        <div className="flex-1" />
        <div className="relative">
          <Search size={12} className="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search alarms..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-44 h-7 text-xs pl-7"
          />
        </div>
      </div>

      <Card>
        <CardHeader className="py-2 px-3 min-h-0">
          <CardTitle>{showActive ? "Active Alarms" : "Alarm History"}</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {filtered.length === 0 && !error ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <AlertTriangle size={28} className="text-muted-foreground/30 mb-2" />
              <p className="text-sm text-muted-foreground">
                {search || severityFilter !== "all"
                  ? "No alarms match your filters."
                  : showActive
                    ? "No active alarms."
                    : "No alarm history recorded."}
              </p>
            </div>
          ) : (
            <div className="overflow-auto max-h-[70vh]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted sticky top-0">
                    <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Severity</th>
                    <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Message</th>
                    <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">PLC</th>
                    <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Time</th>
                    <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Status</th>
                    <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((a) => (
                    <tr key={a.id} className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="px-3 py-1.5 whitespace-nowrap">
                        <Badge variant={sevVariant(a.severity)} className="text-[10px]">{a.severity}</Badge>
                      </td>
                      <td className="px-3 py-1.5 text-sm">{a.message}</td>
                      <td className="px-3 py-1.5 font-mono text-xs">{a.plc_id}</td>
                      <td className="px-3 py-1.5 text-xs whitespace-nowrap text-muted-foreground">
                        {new Date(a.created_at).toLocaleString()}
                      </td>
                      <td className="px-3 py-1.5">
                        <Badge variant={a.active ? "destructive" : "outline"} className="text-[10px]">
                          {a.active ? "Active" : a.acknowledged_at ? "Ack" : "Resolved"}
                        </Badge>
                      </td>
                      <td className="px-3 py-1.5 text-right">
                        {a.active && (
                          <Button size="sm" variant="outline" className="h-6 text-[10px] px-2" onClick={() => acknowledge(a.id)}>
                            Acknowledge
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </main>
  );
}
