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

function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime();
  const sec = Math.floor(diff / 1000);
  const min = Math.floor(sec / 60);

  if (sec < 5) return `${sec} sec ago`;
  if (sec < 60) return `${sec} sec ago`;
  if (min < 60) return `${min} min ago`;
  const hours = Math.floor(min / 60);
  if (hours < 24) return `${hours} hr ago`;
  return `${Math.floor(hours / 24)} day ago`;
}

export default function MachineSummaryCard({ machine }: { machine: Machine }) {
  return (
    <Card>
      <CardContent className="p-4 space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">Machine Name</span>
          <span className="font-medium">{machine.machine_name}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">PLC Make</span>
          <span>{machine.plc_make}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">PLC Model</span>
          <span>{machine.plc_model}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Driver</span>
          <span>{/* No driver info in current model */}—</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">IP Address</span>
          <span>{/* No IP in current model */}—</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Connection Status</span>
          <span>{machine.status}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Collection Status</span>
          <span>{machine.collection_status}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Configured Tags</span>
          <span>{machine.configured_tags}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Enabled Tags</span>
          <span>{machine.enabled_tags}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">Latest Sample</span>
          <span>{machine.last_sample ? relativeTime(machine.last_sample) : "No telemetry"}</span>
        </div>
      </CardContent>
    </Card>
  );
}
