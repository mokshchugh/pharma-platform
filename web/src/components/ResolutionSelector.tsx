const RESOLUTIONS = ["1m", "1h", "1d", "1w"] as const;

export default function ResolutionSelector({
  value,
  onChange,
}: {
  value: string;
  onChange: (r: string) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-xs text-muted-foreground">Resolution</span>
      <div className="flex rounded-sm border border-input overflow-hidden">
        {RESOLUTIONS.map((r) => (
          <button
            key={r}
            onClick={() => onChange(r)}
            className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
              value === r
                ? "bg-primary text-primary-foreground"
                : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            {r}
          </button>
        ))}
      </div>
    </div>
  );
}
