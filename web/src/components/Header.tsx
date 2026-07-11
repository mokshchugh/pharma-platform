interface HeaderProps {
  onToggleSidebar?: () => void;
  sidebarOpen?: boolean;
}

export default function Header({ onToggleSidebar, sidebarOpen }: HeaderProps) {
  return (
    <header className="flex h-12 items-center border-b border-border bg-card px-3 shrink-0 gap-3">
      <button
        onClick={onToggleSidebar}
        className="flex items-center justify-center w-6 h-6 rounded-sm text-muted-foreground hover:text-foreground hover:bg-accent transition-colors cursor-pointer border-none bg-transparent"
        title={sidebarOpen ? "Collapse sidebar" : "Expand sidebar"}
      >
        {sidebarOpen ? "\u2039" : "\u203A"}
      </button>
      <span className="font-semibold text-sm tracking-wide">Pharma Platform</span>
    </header>
  );
}
