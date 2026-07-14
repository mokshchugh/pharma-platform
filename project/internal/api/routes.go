package api

import (
	"github.com/go-chi/chi/v5"

	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/web"
)

type Handlers struct {
	Telemetry  *handlers.TelemetryHandler
	PLC        *handlers.PLCHandler
	Tag        *handlers.TagHandler
	Collector  *handlers.CollectorHandler
	Alarms     *handlers.AlarmHandler
	System     *handlers.SystemHandler
	Machine    *handlers.MachineHandler
	Analytics  *handlers.AnalyticsHandler
	Dashboard  *handlers.DashboardHandler
	OEE        *handlers.OEEHandler
	Production *handlers.ProductionHandler
	Controls   *handlers.ControlHandler
	BizAnalytics *handlers.BusinessAnalyticsHandler
}

func registerAnalyticsRoutes(r chi.Router, h *Handlers) {
	r.Route("/api/v2/analytics", func(r chi.Router) {
		r.Get("/overview", h.BizAnalytics.Overview)
		r.Get("/production", h.BizAnalytics.Production)
		r.Get("/quality", h.BizAnalytics.Quality)
		r.Get("/machines", h.BizAnalytics.Machines)
		r.Get("/energy", h.BizAnalytics.Energy)
		r.Get("/alarms", h.BizAnalytics.Alarms)
		r.Get("/correlations", h.BizAnalytics.Correlations)
		r.Get("/maintenance", h.BizAnalytics.Maintenance)
		r.Get("/insights", h.BizAnalytics.Insights)
	})
}

func Routes(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get("/system/status", h.System.Status)

	r.Get("/telemetry/latest", h.Telemetry.Latest)
	r.Get("/telemetry/latest/{plc_id}", h.Telemetry.LatestByPLC)
	r.Get("/telemetry/latest/{plc_id}/{tag_id}", h.Telemetry.LatestByPLCAndTag)
	r.Get("/telemetry/history", h.Telemetry.History)
	r.Get("/telemetry/aggregate", h.Telemetry.Aggregate)
	r.Get("/telemetry/stream", h.Telemetry.DataStream)
	r.Get("/telemetry/stream/csv", h.Telemetry.DataStreamCSV)

	r.Get("/plcs", h.PLC.List)
	r.Get("/plcs/{plc_id}", h.PLC.Get)
	r.Get("/plcs/{plc_id}/status", h.PLC.GetStatus)
	r.Get("/plcs/{plc_id}/tags", h.PLC.ListTags)
	r.Post("/plcs/{plc_id}/toggle", h.PLC.ToggleEnabled)

	r.Get("/tags", h.Tag.List)
	r.Get("/tags/{tag_id}", h.Tag.Get)

	r.Get("/alarms", h.Alarms.List)
	r.Get("/alarms/active", h.Alarms.ListActive)
	r.Post("/alarms/acknowledge/{id}", h.Alarms.Acknowledge)

	r.Get("/collector/status", h.Collector.Status)
	r.Post("/collector/pause", h.Collector.Pause)
	r.Post("/collector/resume", h.Collector.Resume)

	r.Get("/api/v1/machines", h.Machine.List)
	r.Get("/api/v1/machines/{id}", h.Machine.Get)
	r.Get("/api/v1/machines/{id}/telemetry", h.Analytics.GetTelemetry)
	r.Get("/api/v1/machines/{id}/analytics", h.Analytics.GetAnalytics)

	r.Get("/api/v1/dashboard", h.Dashboard.Summary)

	r.Get("/api/v1/oee", h.OEE.List)
	r.Get("/api/v1/oee/{id}", h.OEE.Get)

	r.Get("/api/v1/production", h.Production.List)
	r.Get("/api/v1/production/active/{machine_id}", h.Production.GetActive)
	r.Post("/api/v1/production/start", h.Production.StartRun)
	r.Post("/api/v1/production/complete/{id}", h.Production.CompleteRun)

	r.Get("/api/v1/downtime", h.Production.ListDowntime)
	r.Post("/api/v1/downtime/start", h.Production.StartDowntime)
	r.Post("/api/v1/downtime/end/{id}", h.Production.EndDowntime)

	r.Get("/api/v1/controls", h.Controls.List)
	r.Get("/api/v1/controls/{id}", h.Controls.Get)
	r.Post("/api/v1/controls/{id}/start", h.Controls.Start)
	r.Post("/api/v1/controls/{id}/stop", h.Controls.Stop)
	r.Post("/api/v1/controls/{id}/setpoint", h.Controls.Setpoint)
	r.Post("/api/v1/controls/{id}/mode", h.Controls.SetMode)

	registerAnalyticsRoutes(r, h)

	r.Handle("/*", web.Handler())

	return r
}

func RoutesBackend(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get("/system/status", h.System.Status)

	r.Get("/telemetry/latest", h.Telemetry.Latest)
	r.Get("/telemetry/latest/{plc_id}", h.Telemetry.LatestByPLC)
	r.Get("/telemetry/latest/{plc_id}/{tag_id}", h.Telemetry.LatestByPLCAndTag)
	r.Get("/telemetry/history", h.Telemetry.History)
	r.Get("/telemetry/aggregate", h.Telemetry.Aggregate)
	r.Get("/telemetry/stream", h.Telemetry.DataStream)
	r.Get("/telemetry/stream/csv", h.Telemetry.DataStreamCSV)

	r.Get("/plcs", h.PLC.List)
	r.Get("/plcs/{plc_id}", h.PLC.Get)
	r.Get("/plcs/{plc_id}/status", h.PLC.GetStatus)
	r.Get("/plcs/{plc_id}/tags", h.PLC.ListTags)
	r.Post("/plcs/{plc_id}/toggle", h.PLC.ToggleEnabled)

	r.Get("/tags", h.Tag.List)
	r.Get("/tags/{tag_id}", h.Tag.Get)

	r.Get("/alarms", h.Alarms.List)
	r.Get("/alarms/active", h.Alarms.ListActive)
	r.Post("/alarms/acknowledge/{id}", h.Alarms.Acknowledge)

	r.Get("/collector/status", h.Collector.Status)
	r.Post("/collector/pause", h.Collector.Pause)
	r.Post("/collector/resume", h.Collector.Resume)

	r.Get("/api/v1/machines", h.Machine.List)
	r.Get("/api/v1/machines/{id}", h.Machine.Get)
	r.Get("/api/v1/machines/{id}/telemetry", h.Analytics.GetTelemetry)
	r.Get("/api/v1/machines/{id}/analytics", h.Analytics.GetAnalytics)

	r.Get("/api/v1/dashboard", h.Dashboard.Summary)

	r.Get("/api/v1/oee", h.OEE.List)
	r.Get("/api/v1/oee/{id}", h.OEE.Get)

	r.Get("/api/v1/production", h.Production.List)
	r.Get("/api/v1/production/active/{machine_id}", h.Production.GetActive)
	r.Post("/api/v1/production/start", h.Production.StartRun)
	r.Post("/api/v1/production/complete/{id}", h.Production.CompleteRun)

	r.Get("/api/v1/downtime", h.Production.ListDowntime)
	r.Post("/api/v1/downtime/start", h.Production.StartDowntime)
	r.Post("/api/v1/downtime/end/{id}", h.Production.EndDowntime)

	r.Get("/api/v1/controls", h.Controls.List)
	r.Get("/api/v1/controls/{id}", h.Controls.Get)
	r.Post("/api/v1/controls/{id}/start", h.Controls.Start)
	r.Post("/api/v1/controls/{id}/stop", h.Controls.Stop)
	r.Post("/api/v1/controls/{id}/setpoint", h.Controls.Setpoint)
	r.Post("/api/v1/controls/{id}/mode", h.Controls.SetMode)

	registerAnalyticsRoutes(r, h)

	return r
}

func routesTelemetryOnly(telemetry *handlers.TelemetryHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get("/telemetry/latest", telemetry.Latest)
	r.Get("/telemetry/history", telemetry.History)

	return r
}
