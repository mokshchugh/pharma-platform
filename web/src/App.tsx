import { useState } from "react";
import Header from "./components/Header";
import Sidebar from "./components/Sidebar";
import HomePage from "./components/HomePage";
import AnalyticsPage from "./components/AnalyticsPage";
import DataStreamPage from "./components/DataStreamPage";
import MachinesPage from "./components/MachinesPage";
import MachineDetailPage from "./components/MachineDetailPage";
import ProductionPage from "./components/ProductionPage";
import AlarmsPage from "./components/AlarmsPage";
import ManagePLCsPage from "./components/ManagePLCsPage";
import ControlsPage from "./components/ControlsPage";

export default function App() {
  const [page, setPage] = useState("Home");
  const [pageParams, setPageParams] = useState<Record<string, string>>({});
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false);
  const [theme, setTheme] = useState<"light" | "dark">("light");

  function handleNavigate(label: string, params?: Record<string, string>) {
    setPage(label);
    if (params) setPageParams(params);
    setMobileDrawerOpen(false);
  }

  function toggleTheme() {
    setTheme((t) => (t === "light" ? "dark" : "light"));
  }

  function renderPage() {
    switch (page) {
      case "Home":
        return <HomePage onNavigate={handleNavigate} />;
      case "Analytics":
        return <AnalyticsPage />;
      case "Data Stream":
        return <DataStreamPage />;
      case "Machines":
        return <MachinesPage onNavigate={handleNavigate} />;
      case "MachineDetail":
        return (
          <MachineDetailPage
            machineId={pageParams.id ?? ""}
            onBack={() => handleNavigate("Machines")}
          />
        );
      case "Production":
        return <ProductionPage />;
      case "Alarms":
        return <AlarmsPage />;
      case "Manage PLCs":
        return <ManagePLCsPage />;
      case "Controls":
        return <ControlsPage />;
      default:
        return <HomePage onNavigate={handleNavigate} />;
    }
  }

  return (
    <div className="flex flex-col h-dvh">
      <Header
        currentPage={page}
        onHamburgerClick={() => setMobileDrawerOpen((v) => !v)}
        theme={theme}
        onThemeToggle={toggleTheme}
      />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar currentPage={page} onNavigate={handleNavigate} mobileDrawerOpen={mobileDrawerOpen} onMobileClose={() => setMobileDrawerOpen(false)} />
        {renderPage()}
      </div>
    </div>
  );
}
