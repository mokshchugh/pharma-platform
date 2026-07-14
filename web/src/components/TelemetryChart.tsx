import { useRef, useEffect } from "react";
import * as echarts from "echarts";

interface AnalyticsPoint {
  timestamp: string;
  avg_value: number;
  min_value: number;
  max_value: number;
  sample_count: number;
}

export default function TelemetryChart({ series, unit }: { series: AnalyticsPoint[]; unit: string }) {
  const chartRef = useRef<HTMLDivElement>(null);
  const instanceRef = useRef<echarts.ECharts>();

  useEffect(() => {
    if (!chartRef.current) return;

    if (!instanceRef.current) {
      instanceRef.current = echarts.init(chartRef.current);
    }

    const chart = instanceRef.current;

    const timestamps = series.map((s) => new Date(s.timestamp).toLocaleString());
    const avgData = series.map((s) => Number(s.avg_value.toFixed(2)));
    const minData = series.map((s) => Number(s.min_value.toFixed(2)));
    const maxData = series.map((s) => Number(s.max_value.toFixed(2)));

    chart.setOption({
      tooltip: {
        trigger: "axis",
        formatter: (params: { data: number; seriesName: string; axisValue: string }[]) => {
          const idx = params[0]?.dataIndex ?? 0;
          const pt = series[idx];
          if (!pt) return "";
          return [
            `<div class="text-xs"><b>${new Date(pt.timestamp).toLocaleString()}</b></div>`,
            `Average: <b>${pt.avg_value.toFixed(2)}</b> ${unit}`,
            `Minimum: <b>${pt.min_value.toFixed(2)}</b> ${unit}`,
            `Maximum: <b>${pt.max_value.toFixed(2)}</b> ${unit}`,
            `Samples: <b>${pt.sample_count}</b>`,
          ].join("<br/>");
        },
      },
      legend: { show: true, bottom: 0, textStyle: { fontSize: 10 } },
      grid: { left: 50, right: 16, top: 8, bottom: 32 },
      xAxis: {
        type: "category",
        data: timestamps,
        axisLabel: { fontSize: 10, rotate: 45 },
        boundaryGap: false,
      },
      yAxis: {
        type: "value",
        axisLabel: { fontSize: 10 },
        name: unit,
        nameTextStyle: { fontSize: 10 },
      },
      series: [
        {
          name: "Average",
          type: "line",
          data: avgData,
          smooth: true,
          symbol: "none",
          lineStyle: { width: 2 },
        },
        {
          name: "Min",
          type: "line",
          data: minData,
          smooth: true,
          symbol: "none",
          lineStyle: { width: 1, type: "dashed" },
        },
        {
          name: "Max",
          type: "line",
          data: maxData,
          smooth: true,
          symbol: "none",
          lineStyle: { width: 1, type: "dashed" },
        },
      ],
    });

    const observer = new ResizeObserver(() => chart.resize());
    observer.observe(chartRef.current);

    return () => observer.disconnect();
  }, [series, unit]);

  return <div ref={chartRef} className="w-full h-64" />;
}
