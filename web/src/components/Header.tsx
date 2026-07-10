export default function Header() {
  return (
    <header style={{ height: "12dvh", borderBottom: "1px solid #ccc", display: "flex", alignItems: "center", padding: "0 16px" }}>
      <span style={{ fontWeight: 600, fontSize: "1.25rem" }}>Pharma Platform</span>
      <div style={{ marginLeft: "auto" }}>User</div>
    </header>
  );
}
