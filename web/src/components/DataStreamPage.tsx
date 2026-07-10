import { useCallback, useEffect, useState } from "react";

const RESOLUTIONS = ["raw", "1m", "1h", "1d", "1w"] as const;

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
  machine_name: string;
}

interface TagOption {
  id: string;
  name: string;
}

export default function DataStreamPage() {
  const [resolution, setResolution] = useState<string>("1m");
  const [start, setStart] = useState(hoursAgoISO(1));
  const [end, setEnd] = useState(nowISO());
  const [machine, setMachine] = useState("");
  const [tag, setTag] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize] = useState(100);

  const [data, setData] = useState<StreamResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [machineOptions, setMachineOptions] = useState<MachineOption[]>([]);
  const [tagOptions, setTagOptions] = useState<TagOption[]>([]);

  useEffect(() => {
    fetch("/plcs")
      .then((r) => r.json())
      .then((list: MachineOption[]) => setMachineOptions(list))
      .catch(() => {});
    fetch("/tags")
      .then((r) => r.json())
      .then((list: TagOption[]) => setTagOptions(list))
      .catch(() => {});
  }, []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError("");

    const params = new URLSearchParams({
      resolution,
      start: new Date(start).toISOString(),
      end: new Date(end).toISOString(),
      page: String(page),
      page_size: String(pageSize),
    });
    if (machine) params.set("machine", machine);
    if (tag) params.set("plc", tag);

    try {
      const res = await fetch(`/telemetry/stream?${params}`);
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || res.statusText);
      }
      const json: StreamResponse = await res.json();
      setData(json);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }, [resolution, start, end, machine, tag, page, pageSize]);

  const isAggregate = resolution !== "raw";
  const columns = isAggregate
    ? ["timestamp", "machine_name", "tag_name", "min_value", "max_value", "avg_value", "sample_count"]
    : ["timestamp", "machine_name", "tag_name", "value", "quality"];

  const totalPages = data ? Math.ceil(data.total / pageSize) : 1;

  return (
    <main style={{ flex: 1, height: "100%", overflow: "auto", padding: "16px" }}>
      <h2 style={{ marginTop: 0 }}>Data Stream</h2>

      <div style={{ display: "flex", gap: "12px", flexWrap: "wrap", marginBottom: "16px" }}>
        <select value={resolution} onChange={(e) => { setResolution(e.target.value); setPage(1); }}>
          {RESOLUTIONS.map((r) => (
            <option key={r} value={r}>
              {r === "raw" ? "Raw" : `${r}`}
            </option>
          ))}
        </select>

        <label>
          Start{" "}
          <input type="datetime-local" value={start} onChange={(e) => { setStart(e.target.value); setPage(1); }} />
        </label>

        <label>
          End{" "}
          <input type="datetime-local" value={end} onChange={(e) => { setEnd(e.target.value); setPage(1); }} />
        </label>

        <select value={machine} onChange={(e) => { setMachine(e.target.value); setPage(1); }}>
          <option value="">Machine (all)</option>
          {machineOptions.map((m) => (
            <option key={m.id} value={m.id}>{m.machine_name}</option>
          ))}
        </select>

        <select value={tag} onChange={(e) => { setTag(e.target.value); setPage(1); }}>
          <option value="">Tag (all)</option>
          {tagOptions.map((t) => (
            <option key={t.id} value={t.name}>{t.name}</option>
          ))}
        </select>

        <button onClick={fetchData} disabled={loading}>
          {loading ? "Loading..." : "Query"}
        </button>
      </div>

      {error && <p style={{ color: "#c00" }}>{error}</p>}

      {data && (
        <>
          <div style={{ overflowX: "auto" }}>
            <table style={{ borderCollapse: "collapse", fontSize: "0.875rem", width: "100%" }}>
              <thead>
                <tr>
                  {columns.map((col) => (
                    <th key={col} style={{ textAlign: "left", padding: "6px 10px", borderBottom: "1px solid #ccc", whiteSpace: "nowrap" }}>
                      {col}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {data.data.length === 0 ? (
                  <tr>
                    <td colSpan={columns.length} style={{ padding: "16px", textAlign: "center", color: "#999" }}>
                      No data
                    </td>
                  </tr>
                ) : (
                  (data.data as (RawRow | AggregateRow)[]).map((row, i) => (
                    <tr key={i}>
                      {columns.map((col) => (
                        <td key={col} style={{ padding: "4px 10px", borderBottom: "1px solid #eee", whiteSpace: "nowrap" }}>
                          {renderCell(row, col)}
                        </td>
                      ))}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: "8px", marginTop: "12px" }}>
            <button disabled={page <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))}>
              Previous
            </button>
            <span>
              Page {data.page} of {totalPages} ({data.total} rows)
            </span>
            <button disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Next
            </button>
          </div>
        </>
      )}
    </main>
  );
}

function renderCell(row: RawRow | AggregateRow, col: string): string {
  switch (col) {
    case "timestamp":
      return new Date((row as RawRow).timestamp).toLocaleString();
    case "machine_name":
      return (row as RawRow).machine_name;
    case "machine_id":
      return (row as RawRow).machine_id;
    case "tag_name":
      return (row as RawRow).tag_name;
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
    default:
      return "";
  }
}
