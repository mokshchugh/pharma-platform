import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const RESOLUTIONS = ["raw", "1m", "1h", "1d", "1w"] as const;
const LIVE_WINDOWS = [
  { label: "Last 15 Minutes", ms: 15 * 60 * 1000 },
  { label: "Last 1 Hour", ms: 3600000 },
  { label: "Last 4 Hours", ms: 4 * 3600000 },
  { label: "Last 8 Hours", ms: 8 * 3600000 },
  { label: "Last 24 Hours", ms: 24 * 3600000 },
] as const;

type Resolution = (typeof RESOLUTIONS)[number];
type ColumnKey = "timestamp" | "machine_name" | "tag_name" | "value" | "quality" | "min_value" | "max_value" | "avg_value" | "sample_count";

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

interface RawRow {
  timestamp: string;
  machine_id: string;
  machine_name: string;
  tag_name: string;
  value: number;
  quality: number;
}

interface AggregateRow {
  timestamp: string;
  machine_id: string;
  machine_name: string;
  tag_name: string;
  min_value: number;
  max_value: number;
  avg_value: number;
  sample_count: number;
}

interface StreamResponse {
  data: RawRow[] | AggregateRow[];
  total: number;
  page: number;
  page_size: number;
  resolution: string;
}

interface MachineOption {
  id: string;
  machine_id: number;
  machine_name: string;
}

interface TagOption {
  id: string;
  name: string;
}

export default function DataStreamPage() {
  const [mode, setMode] = useState<"live" | "range">("live");
  const [resolution, setResolution] = useState<string>("raw");
  const [pollingPaused, setPollingPaused] = useState(false);
  const [liveWindow, setLiveWindow] = useState(3600000);

  const [draftFrom, setDraftFrom] = useState(hoursAgoISO(1));
  const [draftTo, setDraftTo] = useState(nowISO());
  const [appliedFrom, setAppliedFrom] = useState("");
  const [appliedTo, setAppliedTo] = useState("");

  const [machine, setMachine] = useState("");
  const [tag, setTag] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize] = useState(100);

  const [data, setData] = useState<StreamResponse | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const [machineOptions, setMachineOptions] = useState<MachineOption[]>([]);
  const [tagOptions, setTagOptions] = useState<TagOption[]>([]);

  const [csvDialogOpen, setCsvDialogOpen] = useState(false);
  const [csvMachine, setCsvMachine] = useState("");
  const [csvTag, setCsvTag] = useState("");
  const [csvResolution, setCsvResolution] = useState<Resolution>("raw");
  const [csvFrom, setCsvFrom] = useState(hoursAgoISO(1));
  const [csvTo, setCsvTo] = useState(nowISO());
  const [csvTagOptions, setCsvTagOptions] = useState<TagOption[]>([]);

  const stateRef = useRef({
    mode, resolution, machine, tag, liveWindow, appliedFrom, appliedTo, page, pageSize, pollingPaused,
  });
  const tagRef = useRef(tag);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const prevMachineRef = useRef(machine);

  useEffect(() => {
    stateRef.current = {
      mode, resolution, machine, tag, liveWindow, appliedFrom, appliedTo, page, pageSize, pollingPaused,
    };
    tagRef.current = tag;
  });

  useEffect(() => {
    fetch("/plcs")
      .then((r) => {
        if (!r.ok) throw new Error(`GET /plcs ${r.status}`);
        return r.json() as Promise<MachineOption[]>;
      })
      .then((list) => setMachineOptions(list))
      .catch((e) => console.error("Machine fetch failed:", e));
  }, []);

  useEffect(() => {
    let cancelled = false;

    const url = machine ? `/tags?machine_id=${machine}` : "/tags";

    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`GET ${url} ${r.status}`);
        return r.json() as Promise<TagOption[]>;
      })
      .then((list) => {
        if (cancelled) return;
        const deduped = machine ? list : dedupeTags(list);
        setTagOptions(deduped);
        const currentTag = tagRef.current;
        if (currentTag && !deduped.some((t) => t.name === currentTag)) {
          setTag("");
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setTagOptions([]);
          console.error("Tag fetch failed:", e);
        }
      });

    return () => { cancelled = true; };
  }, [machine]);

  useEffect(() => {
    if (!csvDialogOpen) return;
    let cancelled = false;

    const url = csvMachine ? `/tags?machine_id=${csvMachine}` : "/tags";

    fetch(url)
      .then((r) => {
        if (!r.ok) throw new Error(`GET ${url} ${r.status}`);
        return r.json() as Promise<TagOption[]>;
      })
      .then((list) => {
        if (cancelled) return;
        const deduped = csvMachine ? list : dedupeTags(list);
        setCsvTagOptions(deduped);
        if (csvTag && !deduped.some((t) => t.name === csvTag)) {
          setCsvTag("");
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setCsvTagOptions([]);
          console.error("CSV tag fetch failed:", e);
        }
      });

    return () => { cancelled = true; };
  }, [csvDialogOpen, csvMachine]);

  const doFetch = useCallback(async () => {
    const s = stateRef.current;

    if (s.mode === "range" && !s.appliedFrom) return;

    if (s.mode === "range" && new Date(s.appliedTo).getTime() <= Date.now()) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setError("");
    setLoading(true);

    let start: string;
    let end: string;

    if (s.mode === "live") {
      const now = Date.now();
      end = new Date(now).toISOString();
      start = new Date(now - s.liveWindow).toISOString();
    } else {
      const appliedToTime = new Date(s.appliedTo).getTime();
      const now = Date.now();
      end = new Date(Math.min(appliedToTime, now)).toISOString();
      start = new Date(s.appliedFrom).toISOString();
    }

    const params = new URLSearchParams({
      resolution: s.resolution,
      start,
      end,
      page: String(s.page),
      page_size: String(s.pageSize),
    });
    if (s.machine) params.set("machine", s.machine);
    if (s.tag) params.set("plc", s.tag);

    try {
      const res = await fetch(`/telemetry/stream?${params}`, { signal: controller.signal });
      if (controller.signal.aborted) return;
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || res.statusText);
      }
      const json: StreamResponse = await res.json();
      if (controller.signal.aborted) return;
      setData(json);
    } catch (e: unknown) {
      if (controller.signal.aborted) return;
      setError(e instanceof Error ? e.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    const s = stateRef.current;
    const shouldPoll = !s.pollingPaused && (s.mode === "live" || (s.mode === "range" && s.appliedFrom !== "" && new Date(s.appliedTo).getTime() > Date.now()));

    const machineChanged = machine !== prevMachineRef.current;
    prevMachineRef.current = machine;

    if (!machineChanged) {
      doFetch();
    }

    if (shouldPoll) {
      intervalRef.current = setInterval(doFetch, 1000);
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      abortRef.current?.abort();
      abortRef.current = null;
    };
  }, [mode, resolution, machine, tag, liveWindow, appliedFrom, appliedTo, page, pollingPaused, doFetch]);

  function handleModeChange(newMode: "live" | "range") {
    setMode(newMode);
    setPage(1);
  }

  function handleResolutionChange(r: string) {
    setResolution(r);
    setPage(1);
  }

  function handleCsvResolutionChange(r: Resolution) {
    setCsvResolution(r);
  }

  function handleMachineChange(v: string) {
    setMachine(v);
    setPage(1);
  }

  function handleTagChange(v: string) {
    setTag(v);
    setPage(1);
  }

  function handleLiveWindowChange(v: string) {
    setLiveWindow(Number(v));
    setPage(1);
  }

  function handleApplyRange() {
    setAppliedFrom(new Date(draftFrom).toISOString());
    setAppliedTo(new Date(draftTo).toISOString());
    setPage(1);
  }

  function openCsvDialog() {
    setCsvMachine(machine);
    setCsvTag(tag);
    setCsvResolution(resolution as Resolution);
    if (mode === "live") {
      setCsvTo(nowISO());
      setCsvFrom(hoursAgoISO(liveWindow / 3600000));
    } else {
      setCsvFrom(draftFrom);
      setCsvTo(draftTo);
    }
    setCsvDialogOpen(true);
  }

  function handleCsvDownload() {
    const start = new Date(csvFrom).toISOString();
    const end = new Date(csvTo).toISOString();
    const params = new URLSearchParams({ resolution: csvResolution, start, end });
    if (csvMachine) params.set("machine", csvMachine);
    if (csvTag) params.set("plc", csvTag);
    window.open(`/telemetry/stream/csv?${params}`, "_blank");
    setCsvDialogOpen(false);
  }

  const isAggregate = resolution !== "raw";
  const columns: ColumnKey[] = isAggregate
    ? ["timestamp", "machine_name", "tag_name", "min_value", "max_value", "avg_value", "sample_count"]
    : ["timestamp", "machine_name", "tag_name", "value", "quality"];

  const totalPages = data ? Math.ceil(data.total / pageSize) : 1;

  return (
    <main className="flex-1 h-full overflow-auto p-4">
      <h1 className="text-base font-semibold mb-4">Data Stream</h1>

      <div className="flex items-center gap-3 flex-wrap mb-3">
        <div className="flex rounded-sm border border-input overflow-hidden">
          <button
            onClick={() => handleModeChange("live")}
            className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
              mode === "live"
                ? "bg-primary text-primary-foreground"
                : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            Live
          </button>
          <button
            onClick={() => handleModeChange("range")}
            className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
              mode === "range"
                ? "bg-primary text-primary-foreground"
                : "bg-transparent text-foreground hover:bg-accent"
            }`}
          >
            Time Range
          </button>
        </div>

        <div className="w-px h-6 bg-border" />

        <div className="flex rounded-sm border border-input overflow-hidden">
          {RESOLUTIONS.map((r) => (
            <button
              key={r}
              onClick={() => handleResolutionChange(r)}
              className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
                resolution === r
                  ? "bg-primary text-primary-foreground"
                  : "bg-transparent text-foreground hover:bg-accent"
              }`}
            >
              {r === "raw" ? "Raw" : r}
            </button>
          ))}
        </div>

        <div className="w-px h-6 bg-border" />

        <Button
          variant="outline"
          size="sm"
          onClick={() => setPollingPaused((v) => !v)}
        >
          {pollingPaused ? "Resume" : "Pause"}
        </Button>

        <Button variant="outline" size="sm" onClick={openCsvDialog}>
          Download CSV
        </Button>
      </div>

      <div className="flex items-center gap-3 flex-wrap mb-4">
        <Select value={machine} onValueChange={handleMachineChange}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="Machine (all)" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">Machine (all)</SelectItem>
            {machineOptions.map((m) => (
              <SelectItem key={m.id} value={String(m.machine_id)}>{m.machine_name}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={tag} onValueChange={handleTagChange}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="Tag (all)" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">Tag (all)</SelectItem>
            {tagOptions.map((t) => (
              <SelectItem key={t.id} value={t.name}>{t.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <div className="w-px h-6 bg-border" />

        {mode === "live" ? (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Window</span>
            <Select value={String(liveWindow)} onValueChange={handleLiveWindowChange}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {LIVE_WINDOWS.map((w) => (
                  <SelectItem key={w.ms} value={String(w.ms)}>{w.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        ) : (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">From</span>
            <Input
              type="datetime-local"
              value={draftFrom}
              onChange={(e) => setDraftFrom(e.target.value)}
              className="w-44"
            />
            <span className="text-xs text-muted-foreground">To</span>
            <Input
              type="datetime-local"
              value={draftTo}
              onChange={(e) => setDraftTo(e.target.value)}
              className="w-44"
            />
            <Button onClick={handleApplyRange} size="sm" variant="outline">
              Apply Range
            </Button>
          </div>
        )}
      </div>

      {error && (
        <div className="flex items-center gap-2 mb-2">
          <p className="text-sm text-destructive">{error}</p>
          <Button variant="outline" size="sm" onClick={() => doFetch()}>Retry</Button>
        </div>
      )}

      {loading && !data && (
        <div className="animate-pulse space-y-2">
          <div className="h-7 bg-muted rounded w-full" />
          <div className="h-7 bg-muted rounded w-full" />
          <div className="h-7 bg-muted rounded w-3/4" />
          <div className="h-7 bg-muted rounded w-full" />
          <div className="h-7 bg-muted rounded w-5/6" />
        </div>
      )}

      {data && (
        <>
          <div className="border rounded">
            <Table>
              <TableHeader>
                <TableRow>
                  {columns.map((col) => (
                    <TableHead key={col}>{col}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.data.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={columns.length} className="text-center text-muted-foreground py-6">
                      No data
                    </TableCell>
                  </TableRow>
                ) : (
                  (data.data as (RawRow | AggregateRow)[]).map((row, i) => (
                    <TableRow key={i}>
                      {columns.map((col) => (
                        <TableCell key={col} className="whitespace-nowrap font-mono text-xs">
                          {renderCell(row, col)}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          <div className="flex items-center justify-between mt-3">
            <span className="text-xs text-muted-foreground">
              Page {data.page} of {totalPages} ({data.total} rows)
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}

      {!data && !loading && !error && (
        <p className="text-sm text-muted-foreground">Configure filters and select a time range to view data.</p>
      )}

      {csvDialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-background rounded border shadow-lg w-full max-w-md p-5 mx-4">
            <h2 className="text-sm font-semibold mb-4">Download CSV</h2>

            <div className="space-y-3">
              <div>
                <label className="text-xs text-muted-foreground mb-1 block">Machine</label>
                <Select value={csvMachine} onValueChange={(v) => { setCsvMachine(v); setCsvTag(""); }}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Machine (all)" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">Machine (all)</SelectItem>
                    {machineOptions.map((m) => (
                      <SelectItem key={m.id} value={String(m.machine_id)}>{m.machine_name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="text-xs text-muted-foreground mb-1 block">Tag</label>
                <Select value={csvTag} onValueChange={setCsvTag}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Tag (all)" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">Tag (all)</SelectItem>
                    {csvTagOptions.map((t) => (
                      <SelectItem key={t.id} value={t.name}>{t.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <label className="text-xs text-muted-foreground mb-1 block">Resolution</label>
                <div className="flex rounded-sm border border-input overflow-hidden">
                  {RESOLUTIONS.map((r) => (
                    <button
                      key={r}
                      onClick={() => handleCsvResolutionChange(r)}
                      className={`px-3 py-1 text-xs font-medium transition-colors border-r last:border-r-0 cursor-pointer ${
                        csvResolution === r
                          ? "bg-primary text-primary-foreground"
                          : "bg-transparent text-foreground hover:bg-accent"
                      }`}
                    >
                      {r === "raw" ? "Raw" : r}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <label className="text-xs text-muted-foreground mb-1 block">From</label>
                  <Input type="datetime-local" value={csvFrom} onChange={(e) => setCsvFrom(e.target.value)} />
                </div>
                <div className="flex-1">
                  <label className="text-xs text-muted-foreground mb-1 block">To</label>
                  <Input type="datetime-local" value={csvTo} onChange={(e) => setCsvTo(e.target.value)} />
                </div>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 mt-5">
              <Button variant="outline" size="sm" onClick={() => setCsvDialogOpen(false)}>
                Cancel
              </Button>
              <Button size="sm" onClick={handleCsvDownload}>
                Download
              </Button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

function dedupeTags(tags: TagOption[]): TagOption[] {
  const seen = new Set<string>();
  const result: TagOption[] = [];
  for (const t of tags) {
    if (!seen.has(t.name)) {
      seen.add(t.name);
      result.push(t);
    }
  }
  return result;
}

function renderCell(row: RawRow | AggregateRow, col: ColumnKey): string {
  switch (col) {
    case "timestamp":
      return new Date(row.timestamp).toLocaleString();
    case "machine_name":
      return row.machine_name;
    case "tag_name":
      return row.tag_name;
    case "value":
      return String((row as RawRow).value?.toFixed(4) ?? "");
    case "quality":
      return String((row as RawRow).quality ?? "");
    case "min_value":
      return (row as AggregateRow).min_value?.toFixed(4) ?? "";
    case "max_value":
      return (row as AggregateRow).max_value?.toFixed(4) ?? "";
    case "avg_value":
      return (row as AggregateRow).avg_value?.toFixed(4) ?? "";
    case "sample_count":
      return String((row as AggregateRow).sample_count ?? "");
  }
}
