const SIDEBAR_WIDTH = "25%";
const ITEMS = ["Home", "Machines", "Data Stream", "Alarms", "Manage PLCs", "Controls"];

export default function Sidebar({ onNavigate }: { onNavigate?: (page: string) => void }) {
  return (
    <aside
      style={{
        width: SIDEBAR_WIDTH,
        minWidth: 200,
        maxWidth: 320,
        borderRight: "1px solid #ccc",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        padding: "8px",
      }}
    >
      {ITEMS.map((label) => (
        <button
          key={label}
          onClick={() => onNavigate?.(label)}
          style={{
            display: "block",
            width: "100%",
            padding: "10px 12px",
            marginBottom: "4px",
            textAlign: "left",
            border: "none",
            background: "none",
            cursor: "pointer",
          }}
        >
          {label}
        </button>
      ))}
    </aside>
  );
}
