import { useState } from "react";
import Header from "./components/Header";
import Sidebar from "./components/Sidebar";
import HomePage from "./components/HomePage";
import DataStreamPage from "./components/DataStreamPage";
import MachinesPage from "./components/MachinesPage";
import AlarmsPage from "./components/AlarmsPage";
import ManagePLCsPage from "./components/ManagePLCsPage";
import ControlsPage from "./components/ControlsPage";

export default function App() {
  const [page, setPage] = useState("Home");
  const [sidebarOpen, setSidebarOpen] = useState(true);

  function handleNavigate(label: string) {
    setPage(label);
    setSidebarOpen(false);
  }

  return (
    <div className="flex flex-col h-dvh">
      <Header onToggleSidebar={() => setSidebarOpen((v) => !v)} sidebarOpen={sidebarOpen} />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar currentPage={page} onNavigate={handleNavigate} open={sidebarOpen} />
        {page === "Home" && <HomePage />}
        {page === "Data Stream" && <DataStreamPage />}
        {page === "Machines" && <MachinesPage />}
        {page === "Alarms" && <AlarmsPage />}
        {page === "Manage PLCs" && <ManagePLCsPage />}
        {page === "Controls" && <ControlsPage />}
      </div>
    </div>
  );
}
