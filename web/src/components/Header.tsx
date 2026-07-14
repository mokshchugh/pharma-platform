import { Bell, Moon, Sun } from "lucide-react";

const PAGE_LABELS: Record<string, string> = {
  Home: "Dashboard",
  Analytics: "Analytics",
  Machines: "Machines",
  MachineDetail: "Machine Detail",
  "Data Stream": "Data Stream",
  Production: "Production",
  Alarms: "Alarms",
  "Manage PLCs": "Manage PLCs",
  Controls: "Controls",
};

export default function Header({
  currentPage,
  onHamburgerClick,
  theme,
  onThemeToggle,
}: {
  currentPage: string;
  onHamburgerClick?: () => void;
  theme: "light" | "dark";
  onThemeToggle?: () => void;
}) {
  return (
    <header className="flex h-11 items-center border-b border-border bg-card px-3 shrink-0 gap-2 select-none">
      <button
        onClick={onHamburgerClick}
        className="lg:hidden flex items-center justify-center w-7 h-7 text-muted-foreground hover:text-foreground hover:bg-accent transition-colors cursor-pointer border-none bg-transparent rounded"
        title="Open menu"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <line x1="4" y1="6" x2="20" y2="6" />
          <line x1="4" y1="12" x2="20" y2="12" />
          <line x1="4" y1="18" x2="20" y2="18" />
        </svg>
      </button>
      <span className="font-semibold text-sm tracking-wide text-foreground shrink-0">Pharma Platform</span>
      <span className="text-xs text-muted-foreground mx-1">/</span>
      <span className="text-xs text-muted-foreground">{PAGE_LABELS[currentPage] || currentPage}</span>
      <div className="flex-1" />
      <button
        onClick={onThemeToggle}
        className="flex items-center justify-center w-7 h-7 text-muted-foreground hover:text-foreground hover:bg-accent transition-colors cursor-pointer border-none bg-transparent rounded"
        title="Toggle theme"
      >
        {theme === "light" ? <Moon size={14} /> : <Sun size={14} />}
      </button>
      <button
        className="flex items-center justify-center w-7 h-7 text-muted-foreground hover:text-foreground hover:bg-accent transition-colors cursor-pointer border-none bg-transparent rounded"
        title="Notifications"
      >
        <Bell size={14} />
      </button>
      <div className="flex items-center gap-2 pl-1 ml-1 border-l border-border">
        <div className="w-6 h-6 rounded bg-primary/10 text-primary flex items-center justify-center text-xs font-semibold">
          OP
        </div>
      </div>
    </header>
  );
}
