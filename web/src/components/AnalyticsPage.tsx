import { useEffect, useState, useMemo, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import ReactEChartsCore from "echarts-for-react/lib/core";
import * as echarts from "echarts/core";
import { LineChart, BarChart, ScatterChart, PieChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  ToolboxComponent,
  TitleComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";

echarts.use([
  LineChart,
  BarChart,
  ScatterChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  ToolboxComponent,
  TitleComponent,
  CanvasRenderer,
]);

const API_BASE = "/api/v2/analytics";

interface KPIAggregates {
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
}

interface BusinessMetrics {
  machine_id: number;
  machine_name: string;
  running: boolean;
  faulted: boolean;
  total_production: number;
  good_parts: number;
  reject_parts: number;
  production_rate: number;
  running_time_sec: number;
  idle_time_sec: number;
  availability: number;
  performance: number;
  quality: number;
  oee: number;
  utilization: number;
  quality_pct: number;
  reject_pct: number;
  avg_power: number;
  max_power: number;
  energy_per_part: number;
  current: number;
  voltage: number;
  alarm_count: number;
  fault_frequency: number;
  mtbf_hours: number;
  mttr_hours: number;
  temperature: number;
  vibration: number;
  pressure: number;
  cycle_time_sec: number;
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
  machines: BusinessMetrics[];
  aggregates: KPIAggregates;
  generated_at: string;
  simulated: boolean;
}

interface ProductionAnalytics {
  hourly: Record<string, number>;
  daily: Record<string, number>;
  weekly: Record<string, number>;
  per_machine: Record<string, number>;
  shift: Record<string, number>;
  batch: Record<string, number>;
}

interface QualityAnalytics {
  quality_pct: number;
  reject_pct: number;
  first_pass_yield: number;
  reject_trend: { timestamp: string; value: number }[];
  pareto: Record<string, number>;
  per_machine: Record<string, number>;
}

interface TimeSeriesPoint {
  timestamp: string;
  value: number;
}

interface EnergyAnalytics {
  avg_power: number;
  max_power: number;
  current: number;
  voltage: number;
  power_trend: TimeSeriesPoint[];
  energy_trend: TimeSeriesPoint[];
  energy_per_part: number;
  peak_demand: number;
  per_machine: Record<string, BusinessMetrics>;
}

interface AlarmAnalytics {
  active_count: number;
  total_count: number;
  trend: TimeSeriesPoint[];
  per_machine: Record<string, number>;
  critical_trend: TimeSeriesPoint[];
  fault_frequency: number;
}

interface CorrelationPair {
  label: string;
  x: string;
  y: string;
  correlation: number;
  available: boolean;
}

interface CorrelationAnalysis {
  pairs: CorrelationPair[];
  heatmap: Record<string, Record<string, number>>;
  available: boolean;
}

interface MaintenanceRecommendation {
  machine_id: number;
  machine_name: string;
  issue: string;
  severity: string;
  recommendation: string;
  metric: string;
  value: number;
  threshold: number;
}

interface MaintenanceAnalysis {
  recommendations: MaintenanceRecommendation[];
  high_temperature: MaintenanceRecommendation[];
  high_vibration: MaintenanceRecommendation[];
  high_current: MaintenanceRecommendation[];
  frequent_faults: MaintenanceRecommendation[];
  outliers: MaintenanceRecommendation[];
}

interface Insight {
  category: string;
  message: string;
  severity: string;
  metric: string;
  value: string;
}

interface InsightsAnalysis {
  insights: Insight[];
  top_observations: Insight[];
  production_bottlenecks: Insight[];
  underperforming_machines: Insight[];
  high_energy_consumers: Insight[];
  quality_concerns: Insight[];
  operational_recommendations: Insight[];
  business_recommendations: Insight[];
}

function SkeletonBlock({ className }: { className?: string }) {
  return <div className={`animate-pulse bg-secondary rounded ${className ?? ""}`} />;
}

function SectionTitle({ title, count }: { title: string; count?: number }) {
  return (
    <div className="flex items-center gap-2 mb-3">
      <h2 className="text-sm font-semibold text-foreground">{title}</h2>
      {count !== undefined && (
        <span className="text-[11px] text-muted-foreground font-mono">({count})</span>
      )}
    </div>
  );
}

function SimBadge({ simulated }: { simulated?: boolean }) {
  if (!simulated) return null;
  return (
    <Badge variant="outline" className="text-[10px] uppercase tracking-wider text-yellow-500 border-yellow-500/30 ml-auto">
      Simulation
    </Badge>
  );
}

function KpiCard({
  label,
  value,
  unit,
  status,
  formula,
  missing,
  simulated,
}: {
  label: string;
  value?: string | number;
  unit?: string;
  status?: "success" | "warning" | "critical" | "neutral";
  formula?: string;
  missing?: string[];
  simulated?: boolean;
}) {
  const statusDot = status
    ? { success: "bg-green-500", warning: "bg-yellow-500", critical: "bg-red-500", neutral: "bg-gray-400" }[status]
    : undefined;

  return (
    <Card className="relative">
      <CardContent className="p-3 space-y-1">
        {statusDot && <span className={`absolute top-2 right-2 w-1.5 h-1.5 rounded-full ${statusDot}`} />}
        <div className="flex items-center gap-2">
          <div className="text-[10px] text-muted-foreground uppercase tracking-wider font-medium">{label}</div>
          {simulated && <SimBadge simulated />}
        </div>
        {value !== undefined ? (
          <div className="text-lg font-bold">
            {value}
            {unit && <span className="text-xs font-normal text-muted-foreground ml-1">{unit}</span>}
          </div>
        ) : (
          <div className="text-sm text-muted-foreground italic">Unavailable</div>
        )}
        {formula && <div className="text-[10px] text-muted-foreground/60 font-mono">{formula}</div>}
        {missing && missing.length > 0 && (
          <div className="text-[10px] text-yellow-500/80 mt-1">
            Requires: {missing.join(", ")}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function ChartCard({
  title,
  option,
  height = 250,
  simulated,
  empty,
  emptyMessage,
}: {
  title: string;
  option?: Record<string, unknown> | null;
  height?: number;
  simulated?: boolean;
  empty?: boolean;
  emptyMessage?: string;
}) {
  if (empty || !option) {
    return (
      <Card>
        <CardHeader className="py-2 px-3 flex flex-row items-center">
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="p-4 flex items-center justify-center" style={{ height }}>
          <p className="text-sm text-muted-foreground italic">{emptyMessage ?? "No data available"}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="py-2 px-3 flex flex-row items-center">
        <CardTitle>{title}</CardTitle>
        <SimBadge simulated={simulated} />
      </CardHeader>
      <CardContent className="p-2">
        <ReactEChartsCore
          echarts={echarts}
          option={option}
          style={{ height, width: "100%" }}
          notMerge
          lazyUpdate
          opts={{ renderer: "canvas" }}
        />
      </CardContent>
    </Card>
  );
}

function useFetch<T>(url: string | null) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(() => {
    if (!url) {
      setLoading(false);
      return;
    }
    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json() as Promise<T>;
      })
      .then((d) => {
        setData(d);
        setError(null);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [url]);

  useEffect(() => {
    setLoading(true);
    fetchData();
    const id = setInterval(fetchData, 10000);
    return () => clearInterval(id);
  }, [fetchData]);

  return { data, error, loading, refetch: fetchData };
}

function chartTheme() {
  return {
    backgroundColor: "transparent",
    textStyle: { color: "#8b8fa3", fontSize: 11, fontFamily: "Inter, system-ui, sans-serif" },
    grid: { top: 30, right: 16, bottom: 28, left: 48 },
    tooltip: {
      trigger: "axis" as const,
      backgroundColor: "#1a1d27",
      borderColor: "#2a2d3a",
      textStyle: { color: "#e2e4e8", fontSize: 11 },
    },
    legend: {
      textStyle: { color: "#8b8fa3", fontSize: 11 },
      icon: "roundRect",
      itemWidth: 8,
      itemHeight: 8,
    },
    toolbox: {
      feature: {
        saveAsImage: { title: "Export" },
        dataZoom: { title: { zoom: "Zoom", back: "Reset" } },
      },
      iconStyle: { borderColor: "#8b8fa3" },
    },
    dataZoom: [
      { type: "inside" as const, start: 0, end: 100 },
      { type: "slider" as const, start: 0, end: 100, height: 16, bottom: 4, borderColor: "#2a2d3a" },
    ],
    xAxis: {
      type: "category" as const,
      axisLine: { lineStyle: { color: "#2a2d3a" } },
      axisLabel: { color: "#8b8fa3", fontSize: 10 },
      splitLine: { show: false },
    },
    yAxis: {
      type: "value" as const,
      axisLine: { show: false },
      axisLabel: { color: "#8b8fa3", fontSize: 10 },
      splitLine: { lineStyle: { color: "#2a2d3a", type: "dashed" as const } },
    },
    series: [],
  };
}

function trendLine(data: { timestamp: string; value: number }[], color = "#2563eb", name?: string) {
  return {
    type: "line" as const,
    data: data.map((d) => [d.timestamp, d.value]),
    smooth: true,
    symbol: "circle",
    symbolSize: 4,
    lineStyle: { color, width: 2 },
    itemStyle: { color },
    areaStyle: {
      color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: `${color}40` },
        { offset: 1, color: `${color}05` },
      ]),
    },
    name,
  };
}

function barSeries(data: [string, number][], color = "#2563eb") {
  return {
    type: "bar" as const,
    data,
    itemStyle: { color, borderRadius: [2, 2, 0, 0] as [number, number, number, number] },
  };
}

function SeverityBadge({ severity }: { severity: string }) {
  const map: Record<string, string> = {
    critical: "bg-red-500/20 text-red-400 border-red-500/30",
    warning: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    success: "bg-green-500/20 text-green-400 border-green-500/30",
    info: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  };
  return (
    <span className={`text-[10px] uppercase tracking-wider px-1.5 py-0.5 border ${map[severity] ?? "bg-gray-500/20 text-gray-400 border-gray-500/30"}`}>
      {severity}
    </span>
  );
}

export default function AnalyticsPage() {
  const overview = useFetch<ExecutiveOverview>(`${API_BASE}/overview`);
  const production = useFetch<ProductionAnalytics>(`${API_BASE}/production`);
  const quality = useFetch<QualityAnalytics>(`${API_BASE}/quality`);
  const machines = useFetch<Record<string, BusinessMetrics> & { top_utilized: BusinessMetrics[]; least_utilized: BusinessMetrics[] }>(`${API_BASE}/machines`);
  const energy = useFetch<EnergyAnalytics>(`${API_BASE}/energy`);
  const alarms = useFetch<AlarmAnalytics>(`${API_BASE}/alarms`);
  const correlations = useFetch<CorrelationAnalysis>(`${API_BASE}/correlations`);
  const maintenance = useFetch<MaintenanceAnalysis>(`${API_BASE}/maintenance`);
  const insights = useFetch<InsightsAnalysis>(`${API_BASE}/insights`);

  const loading = overview.loading || production.loading || quality.loading;
  const sim = overview.data?.simulated;

  const prodBarOption = useMemo(() => {
    if (!production.data?.hourly) return null;
    const entries = Object.entries(production.data.hourly).sort();
    return {
      ...chartTheme(),
      title: { text: "Hourly Production", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 32, left: 48 },
      series: [barSeries(entries, "#2563eb")],
      tooltip: { ...chartTheme().tooltip, formatter: (p: { value: number[] }) => `${p.value[0]}: ${Number(p.value[1]).toFixed(0)} units` },
    };
  }, [production.data]);

  const prodMachineOption = useMemo(() => {
    if (!production.data?.per_machine) return null;
    const entries = Object.entries(production.data.per_machine).sort((a, b) => b[1] - a[1]);
    return {
      ...chartTheme(),
      title: { text: "Production by Machine", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 32, left: 80 },
      series: [barSeries(entries, "#10b981")],
      xAxis: { ...chartTheme().xAxis, type: "value" as const },
      yAxis: { ...chartTheme().yAxis, type: "category" as const, data: entries.map((e) => e[0]) },
    };
  }, [production.data]);

  const shiftOption = useMemo(() => {
    if (!production.data?.shift) return null;
    const entries = Object.entries(production.data.shift);
    return {
      ...chartTheme(),
      title: { text: "Shift Production", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 28, left: 48 },
      series: [barSeries(entries, "#8b5cf6")],
    };
  }, [production.data]);

  const rejectTrendOption = useMemo(() => {
    if (!quality.data?.reject_trend?.length) return null;
    return {
      ...chartTheme(),
      title: { text: "Reject Trend (24h)", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      series: [trendLine(quality.data.reject_trend, "#ef4444", "Reject Rate")],
    };
  }, [quality.data]);

  const paretoOption = useMemo(() => {
    if (!quality.data?.pareto) return null;
    const entries = Object.entries(quality.data.pareto)
      .filter(([, v]) => v > 0)
      .sort((a, b) => b[1] - a[1]);
    if (!entries.length) return null;
    return {
      ...chartTheme(),
      title: { text: "Reject Pareto", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 36, left: 48 },
      series: [barSeries(entries, "#f59e0b")],
      xAxis: { ...chartTheme().xAxis, axisLabel: { ...chartTheme().xAxis.axisLabel, rotate: 20 } },
    };
  }, [quality.data]);

  const powerTrendOption = useMemo(() => {
    if (!energy.data?.power_trend?.length) return null;
    return {
      ...chartTheme(),
      title: { text: "Power Trend (24h)", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      series: [trendLine(energy.data.power_trend, "#2563eb", "Power (kW)")],
    };
  }, [energy.data]);

  const energyTrendOption = useMemo(() => {
    if (!energy.data?.energy_trend?.length) return null;
    return {
      ...chartTheme(),
      title: { text: "Energy Consumption (24h)", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      series: [trendLine(energy.data.energy_trend, "#10b981", "Energy (kWh)")],
    };
  }, [energy.data]);

  const alarmTrendOption = useMemo(() => {
    if (!alarms.data?.trend?.length) return null;
    return {
      ...chartTheme(),
      title: { text: "Alarm Trend (24h)", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      series: [trendLine(alarms.data.trend, "#ef4444", "Alarms")],
    };
  }, [alarms.data]);

  const alarmMachineOption = useMemo(() => {
    if (!alarms.data?.per_machine) return null;
    const entries = Object.entries(alarms.data.per_machine).sort((a, b) => b[1] - a[1]);
    return {
      ...chartTheme(),
      title: { text: "Alarms by Machine", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 28, left: 80 },
      series: [barSeries(entries, "#dc2626")],
      xAxis: { ...chartTheme().xAxis, type: "value" as const },
      yAxis: { ...chartTheme().yAxis, type: "category" as const, data: entries.map((e) => e[0]) },
    };
  }, [alarms.data]);

  const machineTempOption = useMemo(() => {
    if (!machines.data?.per_machine) return null;
    const entries = Object.values(machines.data.per_machine).map((m) => [m.machine_name, m.temperature] as [string, number]);
    entries.sort((a, b) => b[1] - a[1]);
    return {
      ...chartTheme(),
      title: { text: "Temperature by Machine", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 32, left: 80 },
      series: [barSeries(entries, "#f59e0b")],
      xAxis: { ...chartTheme().xAxis, type: "value" as const },
      yAxis: { ...chartTheme().yAxis, type: "category" as const, data: entries.map((e) => e[0]) },
    };
  }, [machines.data]);

  const machineVibOption = useMemo(() => {
    if (!machines.data?.per_machine) return null;
    const entries = Object.values(machines.data.per_machine).map((m) => [m.machine_name, m.vibration] as [string, number]);
    entries.sort((a, b) => b[1] - a[1]);
    return {
      ...chartTheme(),
      title: { text: "Vibration by Machine", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 16, bottom: 32, left: 80 },
      series: [barSeries(entries, "#ef4444")],
      xAxis: { ...chartTheme().xAxis, type: "value" as const },
      yAxis: { ...chartTheme().yAxis, type: "category" as const, data: entries.map((e) => e[0]) },
    };
  }, [machines.data]);

  const corrOption = useMemo(() => {
    if (!correlations.data?.pairs?.length) return null;
    const pairs = correlations.data.pairs.filter((p) => p.available);
    if (!pairs.length) return null;
    return {
      ...chartTheme(),
      title: { text: "Correlation Matrix", textStyle: { color: "#e2e4e8", fontSize: 13 } },
      grid: { top: 36, right: 60, bottom: 60, left: 80 },
      xAxis: {
        ...chartTheme().xAxis,
        type: "category" as const,
        data: pairs.map((p) => p.label),
        axisLabel: { rotate: 20, color: "#8b8fa3", fontSize: 9 },
      },
      yAxis: {
        ...chartTheme().yAxis,
        type: "value" as const,
        min: -1,
        max: 1,
        axisLabel: { color: "#8b8fa3", fontSize: 10 },
      },
      series: [
        {
          type: "bar" as const,
          data: pairs.map((p) => ({
            value: p.correlation,
            itemStyle: {
              color: p.correlation >= 0 ? `rgba(37,99,235,${Math.abs(p.correlation)})` : `rgba(239,68,68,${Math.abs(p.correlation)})`,
            },
          })),
          barWidth: 20,
        },
      ],
    };
  }, [correlations.data]);

  if (loading && !overview.data) {
    return (
      <main className="flex-1 h-full overflow-auto p-4 space-y-3">
        <SkeletonBlock className="h-6 w-48 mb-2" />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
          {Array.from({ length: 12 }).map((_, i) => <SkeletonBlock key={i} className="h-20" />)}
        </div>
        <SkeletonBlock className="h-48" />
        <SkeletonBlock className="h-48" />
      </main>
    );
  }

  const agg = overview.data?.aggregates;

  return (
    <main className="flex-1 h-full overflow-auto p-4 space-y-4">
      <div className="flex items-center gap-3">
        <h1 className="text-base font-semibold">Analytics</h1>
        <SimBadge simulated={sim} />
        {overview.data && (
          <span className="text-[10px] text-muted-foreground">
            Generated {new Date(overview.data.generated_at).toLocaleTimeString()}
          </span>
        )}
      </div>

      {/* SECTION 1: Executive KPIs */}
      <div>
        <SectionTitle title="Executive KPIs" count={12} />
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-2">
          <KpiCard label="Total Production" value={agg?.total_production?.toLocaleString() ?? "—"} unit="units" simulated={sim} />
          <KpiCard label="Good Parts" value={agg?.total_good_parts?.toLocaleString() ?? "—"} unit="units" simulated={sim} />
          <KpiCard label="Reject Parts" value={agg?.total_reject_parts?.toLocaleString() ?? "—"} unit="units" simulated={sim} />
          <KpiCard label="Availability" value={agg ? `${agg.avg_availability.toFixed(1)}` : "—"} unit="%" simulated={sim} />
          <KpiCard label="Performance" value={agg ? `${agg.avg_performance.toFixed(1)}` : "—"} unit="%" simulated={sim} />
          <KpiCard label="Quality" value={agg ? `${agg.avg_quality.toFixed(1)}` : "—"} unit="%" simulated={sim} />
          <KpiCard label="OEE" value={agg ? `${agg.avg_oee.toFixed(1)}` : "—"} unit="%" simulated={sim} status={agg && agg.avg_oee > 75 ? "success" : agg && agg.avg_oee > 50 ? "warning" : "critical"} />
          <KpiCard label="Utilization" value={agg ? `${agg.avg_utilization.toFixed(1)}` : "—"} unit="%" simulated={sim} />
          <KpiCard label="Running Time" value={overview.data ? `${overview.data.machines.reduce((s, m) => s + m.running_time_sec, 0).toFixed(0)}` : "—"} unit="s" simulated={sim} />
          <KpiCard label="Idle Time" value={overview.data ? `${overview.data.machines.reduce((s, m) => s + m.idle_time_sec, 0).toFixed(0)}` : "—"} unit="s" simulated={sim} />
          <KpiCard label="MTBF" value={agg ? `${agg.avg_mtbf_hours.toFixed(1)}` : "—"} unit="h" simulated={sim} />
          <KpiCard label="MTTR" value={agg ? `${agg.avg_mttr_hours.toFixed(1)}` : "—"} unit="h" simulated={sim} />
        </div>
      </div>

      {/* SECTION 2: Production Analytics */}
      <div>
        <SectionTitle title="Production Analytics" count={6} />
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          <ChartCard title="Hourly Production" option={prodBarOption} simulated={sim} empty={!prodBarOption} />
          <ChartCard title="Production by Machine" option={prodMachineOption} simulated={sim} empty={!prodMachineOption} />
          <ChartCard title="Shift Production" option={shiftOption} simulated={sim} empty={!shiftOption} />
          <Card>
            <CardHeader className="py-2 px-3"><CardTitle>Daily Production</CardTitle></CardHeader>
            <CardContent className="p-3">
              {production.data?.daily && Object.keys(production.data.daily).length > 0 ? (
                <div className="space-y-1">
                  {Object.entries(production.data.daily).sort().reverse().slice(0, 7).map(([day, val]) => (
                    <div key={day} className="flex justify-between items-center text-sm border-b border-border last:border-0 py-1">
                      <span className="text-muted-foreground text-xs">{day}</span>
                      <span className="font-mono">{val.toFixed(0)}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground italic">Daily aggregation available after 24h of operation</p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* SECTION 3: Quality Analytics */}
      <div>
        <SectionTitle title="Quality Analytics" count={5} />
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          <KpiCard label="Quality" value={quality.data ? `${quality.data.quality_pct.toFixed(1)}` : "—"} unit="%" simulated={sim} status={quality.data && quality.data.quality_pct > 95 ? "success" : quality.data && quality.data.quality_pct > 85 ? "warning" : "critical"} />
          <KpiCard label="Reject Rate" value={quality.data ? `${quality.data.reject_pct.toFixed(1)}` : "—"} unit="%" simulated={sim} />
          <KpiCard label="First Pass Yield" value={quality.data ? `${quality.data.first_pass_yield.toFixed(1)}` : "—"} unit="%" simulated={sim} />
          <KpiCard label="Reject Parts" value={agg?.total_reject_parts?.toFixed(0) ?? "—"} unit="units" simulated={sim} />
          <ChartCard title="Reject Trend" option={rejectTrendOption} simulated={sim} empty={!rejectTrendOption} />
          <ChartCard title="Reject Pareto" option={paretoOption} simulated={sim} empty={!paretoOption} />
        </div>
      </div>

      {/* SECTION 4: Machine Analytics */}
      <div>
        <SectionTitle title="Machine Analytics" count={6} />
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          <ChartCard title="Temperature by Machine" option={machineTempOption} simulated={sim} empty={!machineTempOption} />
          <ChartCard title="Vibration by Machine" option={machineVibOption} simulated={sim} empty={!machineVibOption} />
          <Card>
            <CardHeader className="py-2 px-3"><CardTitle>Top Utilized Machines</CardTitle></CardHeader>
            <CardContent className="p-0">
              {machines.data?.top_utilized?.length ? (
                <table className="w-full text-sm">
                  <thead><tr className="border-b border-border">
                    <th className="text-left px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Machine</th>
                    <th className="text-right px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Utilization</th>
                    <th className="text-right px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">OEE</th>
                  </tr></thead>
                  <tbody>
                    {machines.data.top_utilized.map((m) => (
                      <tr key={m.machine_id} className="border-b border-border last:border-0">
                        <td className="px-3 py-1.5">{m.machine_name}</td>
                        <td className="px-3 py-1.5 text-right font-mono">{(m.utilization * 100).toFixed(1)}%</td>
                        <td className="px-3 py-1.5 text-right font-mono">{(m.oee * 100).toFixed(1)}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : <div className="p-3 text-sm text-muted-foreground italic">No machine data</div>}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="py-2 px-3"><CardTitle>Least Utilized Machines</CardTitle></CardHeader>
            <CardContent className="p-0">
              {machines.data?.least_utilized?.length ? (
                <table className="w-full text-sm">
                  <thead><tr className="border-b border-border">
                    <th className="text-left px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Machine</th>
                    <th className="text-right px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Utilization</th>
                    <th className="text-right px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">OEE</th>
                  </tr></thead>
                  <tbody>
                    {machines.data.least_utilized.map((m) => (
                      <tr key={m.machine_id} className="border-b border-border last:border-0">
                        <td className="px-3 py-1.5">{m.machine_name}</td>
                        <td className="px-3 py-1.5 text-right font-mono">{(m.utilization * 100).toFixed(1)}%</td>
                        <td className="px-3 py-1.5 text-right font-mono">{(m.oee * 100).toFixed(1)}%</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : <div className="p-3 text-sm text-muted-foreground italic">No machine data</div>}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* SECTION 5: Energy Analytics */}
      <div>
        <SectionTitle title="Energy Analytics" count={6} />
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-2">
          <KpiCard label="Avg Power" value={energy.data ? `${energy.data.avg_power.toFixed(1)}` : "—"} unit="kW" simulated={sim} />
          <KpiCard label="Max Power" value={energy.data ? `${energy.data.max_power.toFixed(1)}` : "—"} unit="kW" simulated={sim} />
          <KpiCard label="Current" value={energy.data ? `${energy.data.current.toFixed(1)}` : "—"} unit="A" simulated={sim} />
          <KpiCard label="Voltage" value={energy.data ? `${energy.data.voltage.toFixed(0)}` : "—"} unit="V" simulated={sim} />
          <KpiCard label="Energy/Part" value={energy.data ? `${energy.data.energy_per_part.toFixed(3)}` : "—"} unit="kWh" simulated={sim} />
          <KpiCard label="Peak Demand" value={energy.data ? `${energy.data.peak_demand.toFixed(1)}` : "—"} unit="kW" simulated={sim} />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 mt-3">
          <ChartCard title="Power Trend (24h)" option={powerTrendOption} simulated={sim} empty={!powerTrendOption} />
          <ChartCard title="Energy Consumption (24h)" option={energyTrendOption} simulated={sim} empty={!energyTrendOption} />
        </div>
      </div>

      {/* SECTION 6: Alarms */}
      <div>
        <SectionTitle title="Alarms" count={4} />
        <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
          <KpiCard label="Active Alarms" value={alarms.data?.active_count ?? "—"} status={alarms.data && alarms.data.active_count > 0 ? "warning" : "success"} simulated={sim} />
          <KpiCard label="Total Alarms" value={alarms.data?.total_count ?? "—"} simulated={sim} />
          <KpiCard label="Warning Alarms" value={overview.data?.warning_alarms ?? "—"} simulated={sim} />
          <KpiCard label="Critical Alarms" value={overview.data?.critical_alarms ?? "—"} status={overview.data && overview.data.critical_alarms > 0 ? "critical" : "success"} simulated={sim} />
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 mt-3">
          <ChartCard title="Alarm Trend (24h)" option={alarmTrendOption} simulated={sim} empty={!alarmTrendOption} />
          <ChartCard title="Alarms by Machine" option={alarmMachineOption} simulated={sim} empty={!alarmMachineOption} />
        </div>
      </div>

      {/* SECTION 7: Correlation Analysis */}
      <div>
        <SectionTitle title="Correlation Analysis" count={correlations.data?.pairs?.length ?? 0} />
        {correlations.data?.available ? (
          <>
            <ChartCard title="Correlation Matrix" option={corrOption} simulated={sim} empty={!corrOption} />
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-2 mt-3">
              {correlations.data.pairs.filter((p) => p.available).map((p) => (
                <Card key={p.label}>
                  <CardContent className="p-2 space-y-0.5">
                    <div className="text-[10px] text-muted-foreground uppercase tracking-wider">{p.label}</div>
                    <div className={`text-sm font-mono font-bold ${p.correlation > 0.5 ? "text-blue-400" : p.correlation < -0.5 ? "text-red-400" : "text-muted-foreground"}`}>
                      {p.correlation.toFixed(3)}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </>
        ) : (
          <Card><CardContent className="p-4 text-sm text-muted-foreground italic">Correlation analysis requires at least 2 machines with telemetry data</CardContent></Card>
        )}
      </div>

      {/* SECTION 8: Predictive Maintenance */}
      <div>
        <SectionTitle title="Predictive Maintenance" count={maintenance.data?.recommendations?.length ?? 0} />
        {maintenance.data?.recommendations?.length ? (
          <div className="space-y-2">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2">
              <KpiCard label="High Temp Alerts" value={maintenance.data.high_temperature?.length ?? 0} status={(maintenance.data.high_temperature?.length ?? 0) > 0 ? "warning" : "success"} />
              <KpiCard label="High Vibration Alerts" value={maintenance.data.high_vibration?.length ?? 0} status={(maintenance.data.high_vibration?.length ?? 0) > 0 ? "warning" : "success"} />
              <KpiCard label="High Current Alerts" value={maintenance.data.high_current?.length ?? 0} status={(maintenance.data.high_current?.length ?? 0) > 0 ? "warning" : "success"} />
              <KpiCard label="Outliers" value={maintenance.data.outliers?.length ?? 0} status={(maintenance.data.outliers?.length ?? 0) > 0 ? "critical" : "success"} />
            </div>
            <Card>
              <CardHeader className="py-2 px-3"><CardTitle>Recommendations</CardTitle></CardHeader>
              <CardContent className="p-0">
                <table className="w-full text-sm">
                  <thead><tr className="border-b border-border">
                    <th className="text-left px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Machine</th>
                    <th className="text-left px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Issue</th>
                    <th className="text-left px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Severity</th>
                    <th className="text-left px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Recommendation</th>
                    <th className="text-right px-3 py-1.5 font-medium text-muted-foreground text-[10px] uppercase tracking-wider">Value</th>
                  </tr></thead>
                  <tbody>
                    {maintenance.data.recommendations.map((r, i) => (
                      <tr key={i} className="border-b border-border last:border-0 hover:bg-muted/30">
                        <td className="px-3 py-1.5">{r.machine_name}</td>
                        <td className="px-3 py-1.5">{r.issue}</td>
                        <td className="px-3 py-1.5"><SeverityBadge severity={r.severity} /></td>
                        <td className="px-3 py-1.5 text-muted-foreground text-xs">{r.recommendation}</td>
                        <td className="px-3 py-1.5 text-right font-mono">{r.value.toFixed(1)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          </div>
        ) : (
          <Card><CardContent className="p-4 text-sm text-muted-foreground italic">No maintenance recommendations — all machines operating within normal parameters</CardContent></Card>
        )}
      </div>

      {/* SECTION 9: Insights */}
      <div>
        <SectionTitle title="Insights" count={
          (insights.data ? [
            insights.data.top_observations,
            insights.data.production_bottlenecks,
            insights.data.underperforming_machines,
            insights.data.high_energy_consumers,
            insights.data.quality_concerns,
            insights.data.operational_recommendations,
            insights.data.business_recommendations,
          ].reduce((s, a) => s + a.length, 0) : 0)
        } />
        {insights.data ? (
          <div className="space-y-2">
            {insights.data.top_observations.length > 0 && (
              <Card>
                <CardHeader className="py-2 px-3"><CardTitle>Top Observations</CardTitle></CardHeader>
                <CardContent className="p-3 space-y-2">
                  {insights.data.top_observations.map((obs, i) => (
                    <div key={i} className="flex items-start gap-2 text-sm border-b border-border last:border-0 pb-2">
                      <SeverityBadge severity={obs.severity} />
                      <span>{obs.message}</span>
                      <span className="ml-auto text-xs text-muted-foreground font-mono">{obs.value}</span>
                    </div>
                  ))}
                </CardContent>
              </Card>
            )}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
              {insights.data.production_bottlenecks.length > 0 && (
                <Card><CardHeader className="py-2 px-3"><CardTitle>Production Bottlenecks</CardTitle></CardHeader>
                  <CardContent className="p-3 space-y-1 text-sm">
                    {insights.data.production_bottlenecks.map((b, i) => <div key={i} className="flex items-center gap-2"><SeverityBadge severity={b.severity} /><span>{b.message}</span></div>)}
                  </CardContent></Card>
              )}
              {insights.data.high_energy_consumers.length > 0 && (
                <Card><CardHeader className="py-2 px-3"><CardTitle>High Energy Consumers</CardTitle></CardHeader>
                  <CardContent className="p-3 space-y-1 text-sm">
                    {insights.data.high_energy_consumers.map((e, i) => <div key={i} className="flex items-center gap-2"><SeverityBadge severity={e.severity} /><span>{e.message}</span></div>)}
                  </CardContent></Card>
              )}
              {insights.data.quality_concerns.length > 0 && (
                <Card><CardHeader className="py-2 px-3"><CardTitle>Quality Concerns</CardTitle></CardHeader>
                  <CardContent className="p-3 space-y-1 text-sm">
                    {insights.data.quality_concerns.map((q, i) => <div key={i} className="flex items-center gap-2"><SeverityBadge severity={q.severity} /><span>{q.message}</span></div>)}
                  </CardContent></Card>
              )}
              {insights.data.operational_recommendations.length > 0 && (
                <Card><CardHeader className="py-2 px-3"><CardTitle>Operational Recommendations</CardTitle></CardHeader>
                  <CardContent className="p-3 space-y-1 text-sm">
                    {insights.data.operational_recommendations.map((r, i) => <div key={i} className="flex items-center gap-2"><SeverityBadge severity={r.severity} /><span>{r.message}</span></div>)}
                  </CardContent></Card>
              )}
              {insights.data.business_recommendations.length > 0 && (
                <Card><CardHeader className="py-2 px-3"><CardTitle>Business Recommendations</CardTitle></CardHeader>
                  <CardContent className="p-3 space-y-1 text-sm">
                    {insights.data.business_recommendations.map((r, i) => <div key={i} className="flex items-center gap-2"><SeverityBadge severity={r.severity} /><span>{r.message}</span></div>)}
                  </CardContent></Card>
              )}
            </div>
          </div>
        ) : (
          <Card><CardContent className="p-4 text-sm text-muted-foreground italic">Insights generated after sufficient operational data is collected</CardContent></Card>
        )}
      </div>
    </main>
  );
}
