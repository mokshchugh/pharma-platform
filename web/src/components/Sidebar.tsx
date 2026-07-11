const ITEMS = ["Home", "Machines", "Data Stream", "Alarms", "Manage PLCs", "Controls"];

export default function Sidebar({
  currentPage,
  onNavigate,
  open,
}: {
  currentPage?: string;
  onNavigate?: (page: string) => void;
  open: boolean;
}) {
  return (
    <aside
      className={`border-r border-border bg-card flex flex-col shrink-0 transition-[width] duration-150 ease-in-out overflow-hidden ${
        open ? "w-56 min-w-[200px]" : "w-0 min-w-0"
      }`}
    >
      <nav className="flex flex-col p-1.5 gap-0.5 min-w-56">
        {ITEMS.map((label) => (
          <button
            key={label}
            onClick={() => onNavigate?.(label)}
            data-active={label === currentPage || undefined}
            className="flex items-center h-8 px-3 text-sm text-left rounded-sm transition-colors hover:bg-accent hover:text-accent-foreground data-[active]:bg-accent data-[active]:font-medium cursor-pointer border-none bg-transparent w-full"
          >
            {label}
          </button>
        ))}
      </nav>
    </aside>
  );
}
