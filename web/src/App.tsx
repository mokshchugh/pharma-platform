import { useState } from "react";
import Header from "./components/Header";
import Sidebar from "./components/Sidebar";
import HomePage from "./components/HomePage";
import DataStreamPage from "./components/DataStreamPage";

export default function App() {
  const [page, setPage] = useState("Home");

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100dvh" }}>
      <Header />
      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        <Sidebar onNavigate={setPage} />
        {page === "Home" && <HomePage />}
        {page === "Data Stream" && <DataStreamPage />}
      </div>
    </div>
  );
}
