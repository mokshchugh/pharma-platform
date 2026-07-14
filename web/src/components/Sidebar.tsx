import {
  LayoutDashboard,
  Cpu,
  Radio,
  Factory,
  AlertTriangle,
  Settings,
  Gamepad2,
  BarChart3,
} from "lucide-react";

const ITEMS = [
  { label: "Home", icon: LayoutDashboard },
  { label: "Analytics", icon: BarChart3 },
  { label: "Machines", icon: Cpu },
  { label: "Data Stream", icon: Radio },
  { label: "Production", icon: Factory },
  { label: "Alarms", icon: AlertTriangle },
  { label: "Manage PLCs", icon: Settings },
  { label: "Controls", icon: Gamepad2 },
];

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
    <nav className="flex flex-col py-2 gap-0.5">
      <div className="px-3 pb-2 text-[10px] font-semibold uppercase tracking-widest text-sidebar-foreground/50">
        Navigation
      </div>
      {ITEMS.map(({ label, icon: Icon }) => (
        <button
          key={label}
          onClick={() => {
            onNavigate?.(label);
            onMobileClose();
          }}
          data-active={label === currentPage || undefined}
          className="flex items-center gap-2.5 h-8 px-3 mx-1.5 text-sm text-left rounded transition-colors
            text-sidebar-foreground/80 hover:text-sidebar-foreground hover:bg-sidebar-muted
            data-[active]:bg-sidebar-active data-[active]:text-white
            cursor-pointer border-none bg-transparent w-full"
        >
          <Icon size={15} strokeWidth={1.5} />
          <span>{label === "Home" ? "Dashboard" : label}</span>
        </button>
      ))}
    </nav>
  );

  return (
    <>
      <aside className="hidden lg:flex border-r border-border bg-sidebar text-sidebar-foreground flex-col shrink-0 w-56 overflow-hidden">
        {nav}
      </aside>

      {mobileDrawerOpen && (
        <div className="fixed inset-0 z-40 lg:hidden">
          <div className="absolute inset-0 bg-black/40" onClick={onMobileClose} />
          <aside className="absolute left-0 top-0 bottom-0 w-56 bg-sidebar text-sidebar-foreground border-r border-border flex flex-col z-50 shadow-lg">
            {nav}
          </aside>
        </div>
      )}
    </>
  );
}
