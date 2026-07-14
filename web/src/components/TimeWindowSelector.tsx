import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const PRESETS = [
  { label: "Last Hour", value: "1h" },
  { label: "Last Day", value: "24h" },
  { label: "Last Week", value: "168h" },
] as const;

function nowISO() {
  const d = new Date();
  d.setSeconds(0, 0);
  return d.toISOString().slice(0, 16);
}

function hoursAgoISO(n: number) {
  const d = new Date(Date.now() - n * 3600_000);
  d.setSeconds(0, 0);
  return d.toISOString().slice(0, 16);
}

export default function TimeWindowSelector({
  value,
  onChange,
}: {
  value: string;
  onChange: (preset: string, from?: string, to?: string) => void;
}) {
  const [isCustom, setIsCustom] = useState(value === "custom");
  const [draftFrom, setDraftFrom] = useState(hoursAgoISO(1));
  const [draftTo, setDraftTo] = useState(nowISO());

  function handlePreset(p: string) {
    setIsCustom(false);
    onChange(p);
  }

  function handleCustom() {
    setIsCustom(true);
    onChange("custom", new Date(draftFrom).toISOString(), new Date(draftTo).toISOString());
  }

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <span className="text-xs text-muted-foreground">Time Window</span>
      <div className="flex rounded-sm border border-input overflow-hidden">
        {PRESETS.map((p) => (
          <button
            key={p.value}
            onClick={() => handlePreset(p.value)}
            className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
              !isCustom && value === p.value
                ? "bg-primary text-primary-foreground"
                : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            {p.label}
          </button>
        ))}
        <button
          onClick={() => setIsCustom(true)}
          className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
            isCustom
              ? "bg-primary text-primary-foreground"
              : "bg-transparent text-foreground hover:bg-accent"
          }`}
        >
          Custom
        </button>
      </div>

      {isCustom && (
        <div className="flex items-center gap-2">
          <Input
            type="datetime-local"
            value={draftFrom}
            onChange={(e) => setDraftFrom(e.target.value)}
            className="w-44 h-7 text-xs"
          />
          <span className="text-xs text-muted-foreground">to</span>
          <Input
            type="datetime-local"
            value={draftTo}
            onChange={(e) => setDraftTo(e.target.value)}
            className="w-44 h-7 text-xs"
          />
          <Button variant="outline" size="sm" onClick={handleCustom}>
            Apply
          </Button>
        </div>
      )}
    </div>
  );
}
