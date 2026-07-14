import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import {
  Cpu,
  Radio,
  AlertTriangle,
  Settings,
  Gamepad2,
  Factory,
  BarChart3,
  Activity,
  Clock,
  Zap,
} from "lucide-react";

interface MachineBrief {
  id: number;
  name: string;
  status: string;
  oee_score: number;
}

interface DashboardData {
  total_machines: number;
  running_machines: number;
  stopped_machines: number;
  active_alarms: number;
  critical_alarms: number;
  overall_oee: number;
  today_good_parts: number;
  today_bad_parts: number;
  machine_states: MachineBrief[];
}

interface SystemStatus {
  status: string;
  plcs: { total: number; online: number; offline: number };
  alarms: { active: number; critical: number };
  collector: { status: string };
}

interface ExecutiveOverview {
  plant_status: string;
  collector_status: string;
  questdb_status: string;
  configured_machines: number;
  collecting_machines: number;
  configured_plcs: number;
  configured_tags: number;
  samples_per_sec: number;
  telemetry_today: number;
  latest_sample: string;
  active_alarms: number;
  critical_alarms: number;
  warning_alarms: number;
  machines: Array<{
    machine_id: number;
    machine_name: string;
    running: boolean;
    faulted: boolean;
    oee: number;
    availability: number;
    performance: number;
    quality: number;
    utilization: number;
    running_time_sec: number;
    idle_time_sec: number;
    avg_power: number;
    temperature: number;
  }>;
  aggregates: {
    total_production: number;
    total_good_parts: number;
    total_reject_parts: number;
    avg_availability: number;
    avg_performance: number;
    avg_quality: number;
    avg_oee: number;
    avg_utilization: number;
    total_alarms: number;
    total_critical: number;
    avg_power: number;
    peak_power: number;
    avg_energy_per_part: number;
    avg_mtbf_hours: number;
    avg_mttr_hours: number;
  };
  generated_at: string;
  simulated: boolean;
}

function SkeletonBlock({ className }: { className?: string }) {
  return <div className={`animate-pulse bg-secondary rounded ${className ?? ""}`} />;
}

function StatCard({ label, value, unit, color, icon: Icon }: { label: string; value: string | number; unit?: string; color?: string; icon?: React.ElementType }) {
  return (
    <Card>
      <CardContent className="p-3 space-y-0.5 relative">
        {Icon && <Icon size={14} className="absolute top-2 right-2 text-muted-foreground/20" strokeWidth={1.5} />}
        <div className="text-[11px] text-muted-foreground uppercase tracking-wider font-medium">{label}</div>
        <div className={`text-xl font-bold ${color ?? ""}`}>
          {value}{unit && <span className="text-sm font-normal text-muted-foreground ml-1">{unit}</span>}
        </div>
      </CardContent>
    </Card>
  );
}

function QNCard({ icon: Icon, label, desc, onClick }: { icon: React.ElementType; label: string; desc: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-3 p-3 border border-border bg-card hover:bg-accent transition-colors cursor-pointer text-left w-full"
    >
      <div className="w-9 h-9 flex items-center justify-center bg-primary/10 text-primary shrink-0">
        <Icon size={18} strokeWidth={1.5} />
      </div>
      <div>
        <div className="text-sm font-medium">{label}</div>
        <div className="text-xs text-muted-foreground">{desc}</div>
      </div>
    </button>
  );
}

const QUICK_NAV = [
  { label: "Machines", icon: Cpu, desc: "View machine status and telemetry", target: "Machines" },
  { label: "Analytics", icon: BarChart3, desc: "Production KPIs and insights", target: "Analytics" },
  { label: "Telemetry", icon: Radio, desc: "Browse live and historical data", target: "Data Stream" },
  { label: "Alarms", icon: AlertTriangle, desc: "Active alarm events", target: "Alarms" },
  { label: "Production", icon: Factory, desc: "Production runs and downtime", target: "Production" },
  { label: "Controls", icon: Gamepad2, desc: "Machine operations", target: "Controls" },
];

export default function HomePage({ onNavigate }: { onNavigate?: (label: string) => void }) {
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [system, setSystem] = useState<SystemStatus | null>(null);
  const [biz, setBiz] = useState<ExecutiveOverview | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(() => {
    Promise.all([
      fetch("/api/v1/dashboard").then((r) => { if (!r.ok) throw new Error("Dashboard fetch failed"); return r.json(); }),
      fetch("/system/status").then((r) => { if (!r.ok) throw new Error("System status fetch failed"); return r.json(); }),
      fetch("/api/v2/analytics/overview").then((r) => { if (!r.ok) throw new Error("Analytics fetch failed"); return r.json(); }),
    ])
      .then(([d, s, b]) => {
        setDashboard(d);
        setSystem(s);
        setBiz(b);
        setError("");
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchAll();
    const id = setInterval(fetchAll, 5000);
    return () => clearInterval(id);
  }, [fetchAll]);

  if (loading) {
    return (
      <main className="flex-1 h-full overflow-auto p-4 space-y-3">
        <SkeletonBlock className="h-6 w-32 mb-1" />
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
          {Array.from({ length: 6 }).map((_, i) => <SkeletonBlock key={i} className="h-20" />)}
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          <SkeletonBlock className="h-48" />
          <SkeletonBlock className="h-48" />
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 h-full overflow-auto p-4 space-y-3">
      <h1 className="text-base font-semibold">Dashboard</h1>

      {error && (
        <div className="flex items-center gap-2 px-3 py-2 bg-red-50 border border-red-200 text-red-700 text-sm">
          <span>{error}</span>
          <Button size="sm" variant="outline" className="ml-auto" onClick={fetchAll}>Retry</Button>
        </div>
      )}

      {/* Status badges */}
      <div className="flex items-center gap-2 flex-wrap">
        {system && (
          <>
            <Badge variant={system.status === "ok" ? "default" : system.status === "critical" ? "destructive" : "secondary"} className="text-[11px] uppercase tracking-wider px-2 py-0.5">
              Plant: {system.status === "ok" ? "Healthy" : system.status === "critical" ? "Critical" : "Warning"}
            </Badge>
            <Badge variant={system.collector.status === "running" ? "default" : "secondary"} className="text-[11px] uppercase tracking-wider px-2 py-0.5">
              Collector: {system.collector.status}
            </Badge>
          </>
        )}
        {biz && (
          <Badge variant={biz.questdb_status === "connected" ? "default" : "destructive"} className="text-[11px] uppercase tracking-wider px-2 py-0.5">
            QuestDB: {biz.questdb_status}
          </Badge>
        )}
        {biz?.simulated && (
          <Badge variant="outline" className="text-[11px] uppercase tracking-wider text-yellow-500 border-yellow-500/30">
            Simulation
          </Badge>
        )}
        {biz && (
          <span className="text-[10px] text-muted-foreground">
            Updated {new Date(biz.generated_at).toLocaleTimeString()}
          </span>
        )}
      </div>

      {/* Plant KPIs */}
      {biz && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-2">
          <StatCard label="Configured Machines" value={biz.configured_machines} icon={Cpu} />
          <StatCard label="Collecting" value={biz.collecting_machines} color="text-green-500" icon={Activity} />
          <StatCard label="Configured PLCs" value={biz.configured_plcs} icon={Settings} />
          <StatCard label="Configured Tags" value={biz.configured_tags} icon={Cpu} />
          <StatCard label="Samples/sec" value={biz.samples_per_sec} icon={Radio} />
          <StatCard label="Telemetry Today" value={biz.telemetry_today.toLocaleString()} icon={Clock} />
        </div>
      )}

      {!biz && dashboard && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-2">
          <StatCard label="Total Machines" value={dashboard.total_machines} />
          <StatCard label="Running" value={dashboard.running_machines} color="text-green-600" />
          <StatCard label="Stopped" value={dashboard.stopped_machines} />
          <StatCard label="Active Alarms" value={dashboard.active_alarms} color={dashboard.active_alarms > 0 ? "text-red-600" : "text-green-600"} />
          <StatCard label="Good Parts" value={dashboard.today_good_parts.toLocaleString()} />
          <StatCard label="Overall OEE" value={(dashboard.overall_oee * 100).toFixed(1)} unit="%" />
        </div>
      )}

      {/* KPI Grid */}
      {biz && (
        <>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-2">
            <StatCard label="Total Production" value={biz.aggregates.total_production.toLocaleString()} unit="units" icon={Factory} />
            <StatCard label="Good Parts" value={biz.aggregates.total_good_parts.toLocaleString()} unit="units" icon={Factory} />
            <StatCard label="Reject Parts" value={biz.aggregates.total_reject_parts.toLocaleString()} unit="units" icon={AlertTriangle} />
            <StatCard label="OEE" value={biz.aggregates.avg_oee.toFixed(1)} unit="%" color={biz.aggregates.avg_oee > 75 ? "text-green-500" : biz.aggregates.avg_oee > 50 ? "text-yellow-500" : "text-red-500"} icon={Activity} />
            <StatCard label="Availability" value={biz.aggregates.avg_availability.toFixed(1)} unit="%" icon={Clock} />
            <StatCard label="Performance" value={biz.aggregates.avg_performance.toFixed(1)} unit="%" icon={Activity} />
            <StatCard label="Quality" value={biz.aggregates.avg_quality.toFixed(1)} unit="%" color={biz.aggregates.avg_quality > 95 ? "text-green-500" : "text-yellow-500"} icon={Factory} />
            <StatCard label="Utilization" value={biz.aggregates.avg_utilization.toFixed(1)} unit="%" icon={Clock} />
            <StatCard label="Avg Power" value={biz.aggregates.avg_power.toFixed(1)} unit="kW" icon={Zap} />
            <StatCard label="Peak Power" value={biz.aggregates.peak_power.toFixed(1)} unit="kW" icon={Zap} />
            <StatCard label="MTBF" value={biz.aggregates.avg_mtbf_hours.toFixed(1)} unit="h" icon={Clock} />
            <StatCard label="Energy/Part" value={biz.aggregates.avg_energy_per_part.toFixed(3)} unit="kWh" icon={Zap} />
          </div>

          {/* System + Alarms overview */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
            <Card>
              <CardHeader className="py-2 px-3"><CardTitle>System Overview</CardTitle></CardHeader>
              <CardContent className="p-3 text-sm">
                <div className="grid grid-cols-3 gap-2">
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">PLCs</div>
                    <div className="font-semibold mt-0.5">{system?.plcs.total ?? biz.configured_plcs}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Online</div>
                    <div className="font-semibold mt-0.5 text-green-500">{system?.plcs.online ?? biz.collecting_machines}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Offline</div>
                    <div className="font-semibold mt-0.5 text-red-500">{system?.plcs.offline ?? biz.configured_machines - biz.collecting_machines}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Collector</div>
                    <div className="font-semibold mt-0.5">{system?.collector.status ?? biz.collector_status}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Machines</div>
                    <div className="font-semibold mt-0.5">{biz.configured_machines}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Tags</div>
                    <div className="font-semibold mt-0.5">{biz.configured_tags}</div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="py-2 px-3"><CardTitle>Alarms</CardTitle></CardHeader>
              <CardContent className="p-3 space-y-2 text-sm">
                <div className="grid grid-cols-3 gap-2">
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Active</div>
                    <div className={`font-semibold mt-0.5 ${biz.active_alarms > 0 ? "text-red-500" : "text-green-500"}`}>{biz.active_alarms}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Warning</div>
                    <div className="font-semibold mt-0.5 text-yellow-500">{biz.warning_alarms}</div>
                  </div>
                  <div className="border border-border p-2">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">Critical</div>
                    <div className={`font-semibold mt-0.5 ${biz.critical_alarms > 0 ? "text-red-500" : "text-green-500"}`}>{biz.critical_alarms}</div>
                  </div>
                </div>
                <div className="text-xs text-muted-foreground pt-1">Latest sample: {new Date(biz.latest_sample).toLocaleTimeString()}</div>
              </CardContent>
            </Card>
          </div>
        </>
      )}

      {/* Machine States Table */}
      {biz && biz.machines.length > 0 && (
        <Card>
          <CardHeader className="py-2 px-3"><CardTitle>Machine States</CardTitle></CardHeader>
          <CardContent className="p-0">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Machine</th>
                  <th className="text-left px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Status</th>
                  <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">OEE</th>
                  <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Temp</th>
                  <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Power</th>
                  <th className="text-right px-3 py-2 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Running</th>
                </tr>
              </thead>
              <tbody>
                {biz.machines.map((m) => (
                  <tr key={m.machine_id} className="border-b border-border last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-3 py-1.5 text-sm">{m.machine_name}</td>
                    <td className="px-3 py-1.5">
                      <div className="flex items-center gap-1.5">
                        <span className={`w-2 h-2 rounded-full ${m.faulted ? "bg-red-500" : m.running ? "bg-green-500" : "bg-gray-400"}`} />
                        <span className="text-xs capitalize">{m.faulted ? "Fault" : m.running ? "Running" : "Stopped"}</span>
                      </div>
                    </td>
                    <td className="px-3 py-1.5 text-right font-mono">{(m.oee * 100).toFixed(1)}%</td>
                    <td className="px-3 py-1.5 text-right font-mono">{m.temperature.toFixed(1)}°C</td>
                    <td className="px-3 py-1.5 text-right font-mono">{m.avg_power.toFixed(1)}kW</td>
                    <td className="px-3 py-1.5 text-right font-mono">{(m.running_time_sec / 3600).toFixed(1)}h</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </CardContent>
        </Card>
      )}

      {/* Quick Navigation */}
      <div>
        <div className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Quick Navigation</div>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-2">
          {QUICK_NAV.map((item) => (
            <QNCard key={item.label} icon={item.icon} label={item.label} desc={item.desc} onClick={() => onNavigate?.(item.target)} />
          ))}
        </div>
      </div>
    </main>
  );
}
