const ITEMS = ["Home", "Machines", "Data Stream", "Alarms", "Manage PLCs", "Controls"];

export default function Sidebar({
  currentPage,
  onNavigate,
  mobileDrawerOpen,
  onMobileClose,
}: {
  currentPage?: string;
  onNavigate?: (page: string) => void;
  mobileDrawerOpen: boolean;
  onMobileClose: () => void;
}) {
  const nav = (
    <nav className="flex flex-col p-1.5 gap-0.5 min-w-56">
      {ITEMS.map((label) => (
        <button
          key={label}
          onClick={() => {
            onNavigate?.(label);
            onMobileClose();
          }}
          data-active={label === currentPage || undefined}
          className="flex items-center h-8 px-3 text-sm text-left rounded-sm transition-colors hover:bg-accent hover:text-accent-foreground data-[active]:bg-accent data-[active]:font-medium cursor-pointer border-none bg-transparent w-full"
        >
          {label}
        </button>
      ))}
    </nav>
  );

  return (
    <>
      <aside className="hidden md:flex border-r border-border bg-card flex-col shrink-0 w-56">
        {nav}
      </aside>

      {mobileDrawerOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <div className="absolute inset-0 bg-black/40" onClick={onMobileClose} />
          <aside className="absolute left-0 top-0 bottom-0 w-56 bg-card border-r border-border flex flex-col z-50 shadow-lg">
            {nav}
          </aside>
        </div>
      )}
    </>
  );
}
