export default function Header({ onHamburgerClick }: { onHamburgerClick?: () => void }) {
  return (
    <header className="flex h-12 items-center border-b border-border bg-card px-3 shrink-0 gap-3">
      <button
        onClick={onHamburgerClick}
        className="md:hidden flex items-center justify-center w-6 h-6 rounded-sm text-muted-foreground hover:text-foreground hover:bg-accent transition-colors cursor-pointer border-none bg-transparent"
        title="Open menu"
      >
        &#9776;
      </button>
      <span className="font-semibold text-sm tracking-wide">Pharma Platform</span>
    </header>
  );
}
