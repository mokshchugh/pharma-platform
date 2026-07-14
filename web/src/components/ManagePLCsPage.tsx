import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Settings, Wifi, WifiOff } from "lucide-react";

interface PLC {
  id: string;
  machine_name: string;
  driver: string;
  ip_address: string;
  port: number;
  enabled: boolean;
}

export default function ManagePLCsPage() {
  const [plcs, setPLCs] = useState<PLC[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchPLCs = useCallback(() => {
    setError("");
    fetch("/plcs")
      .then((r) => {
        if (!r.ok) throw new Error(`GET /plcs ${r.status}`);
        return r.json();
      })
      .then((data) => {
        setPLCs(data ?? []);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.message);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchPLCs();
    const id = setInterval(fetchPLCs, 10000);
    return () => clearInterval(id);
  }, [fetchPLCs]);

  const allEnabled = plcs.length > 0 && plcs.every((p) => p.enabled);
  const allDisabled = plcs.length > 0 && plcs.every((p) => !p.enabled);

  const toggleAll = (enable: boolean) => {
    Promise.all(
      plcs.map((p) =>
        fetch(`/plcs/${p.id}/toggle`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ enabled: enable }),
        }).catch(() => {}),
      ),
    ).then(fetchPLCs);
  };

  const toggleOne = (id: string, enabled: boolean) => {
    fetch(`/plcs/${id}/toggle`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ enabled: !enabled }),
    }).then(fetchPLCs);
  };

  if (loading) {
    return (
      <main className="flex-1 h-full overflow-auto p-4 space-y-3">
        <div className="h-5 bg-secondary w-28 animate-pulse" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-28 bg-secondary animate-pulse" />
          ))}
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4 space-y-3">
      <div className="flex items-center justify-between">
        <h1 className="text-base font-semibold">Manage PLCs</h1>
        {plcs.length > 0 && (
          <div className="flex gap-1.5">
            <Button
              size="sm"
              variant="outline"
              className="h-7 text-xs px-2"
              onClick={() => toggleAll(true)}
              disabled={allEnabled}
            >
              Enable All
            </Button>
            <Button
              size="sm"
              variant="outline"
              className="h-7 text-xs px-2"
              onClick={() => toggleAll(false)}
              disabled={allDisabled}
            >
              Disable All
            </Button>
          </div>
        )}
      </div>

      {error && (
        <div className="flex items-center gap-2 px-3 py-2 bg-red-900/30 border border-red-800 text-red-400 text-sm">
          <span>{error}</span>
          <Button size="sm" variant="outline" className="ml-auto" onClick={fetchPLCs}>Retry</Button>
        </div>
      )}

      {!error && plcs.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Settings size={32} className="text-muted-foreground/30 mb-3" />
          <p className="text-sm text-muted-foreground">No PLCs configured.</p>
          <p className="text-xs text-muted-foreground/60 mt-1">PLCs will appear here once they are added to the configuration.</p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2">
        {plcs.map((plc) => (
          <Card key={plc.id}>
            <CardHeader className="flex flex-row items-center justify-between py-2 px-3 min-h-0">
              <div className="flex items-center gap-2 min-w-0">
                {plc.enabled ? (
                  <Wifi size={14} className="text-green-500 shrink-0" />
                ) : (
                  <WifiOff size={14} className="text-muted-foreground shrink-0" />
                )}
                <CardTitle className="text-sm truncate">{plc.machine_name}</CardTitle>
              </div>
              <Badge variant={plc.enabled ? "default" : "secondary"} className="text-[10px] shrink-0 ml-2">
                {plc.enabled ? "Enabled" : "Disabled"}
              </Badge>
            </CardHeader>
            <CardContent className="px-3 pb-3 space-y-1.5 text-xs">
              <div className="grid grid-cols-2 gap-1">
                <span className="text-muted-foreground">ID</span>
                <span className="font-mono text-right truncate">{plc.id}</span>
                <span className="text-muted-foreground">Protocol</span>
                <span className="text-right truncate">{plc.driver}</span>
                <span className="text-muted-foreground">Address</span>
                <span className="font-mono text-right truncate">{plc.ip_address}:{plc.port}</span>
              </div>
              <div className="flex gap-1.5 pt-1.5">
                <Button
                  size="sm"
                  variant={plc.enabled ? "destructive" : "default"}
                  className="h-6 text-[10px] px-2"
                  onClick={() => toggleOne(plc.id, plc.enabled)}
                >
                  {plc.enabled ? "Disable" : "Enable"}
                </Button>
                <Button size="sm" variant="outline" className="h-6 text-[10px] px-2" disabled title="Coming Soon">
                  Configure
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </main>
  );
}
