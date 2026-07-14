import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface AnalyticsPoint {
  timestamp: string;
  avg_value: number;
  min_value: number;
  max_value: number;
  sample_count: number;
}

export default function HistoricalTable({
  series,
  unit,
}: {
  series: AnalyticsPoint[];
  unit: string;
}) {
  return (
    <div className="border rounded overflow-auto max-h-64">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Timestamp</TableHead>
            <TableHead>Telemetry</TableHead>
            <TableHead>Average ({unit})</TableHead>
            <TableHead>Minimum ({unit})</TableHead>
            <TableHead>Maximum ({unit})</TableHead>
            <TableHead>Sample Count</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {series.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground py-6">
                No data
              </TableCell>
            </TableRow>
          ) : (
            series.map((pt, i) => (
              <TableRow key={i}>
                <TableCell className="whitespace-nowrap font-mono text-xs">
                  {new Date(pt.timestamp).toLocaleString()}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {pt.avg_value.toFixed(2)}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {pt.min_value.toFixed(2)}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {pt.max_value.toFixed(2)}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {pt.avg_value.toFixed(2)}
                </TableCell>
                <TableCell className="font-mono text-xs">
                  {pt.sample_count.toLocaleString()}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
