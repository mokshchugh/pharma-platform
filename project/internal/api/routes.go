package api

import (
	"github.com/go-chi/chi/v5"

	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/web"
)

type Handlers struct {
	Telemetry *handlers.TelemetryHandler
	PLC       *handlers.PLCHandler
	Tag       *handlers.TagHandler
	Collector *handlers.CollectorHandler
	Alarms    *handlers.AlarmHandler
	System    *handlers.SystemHandler
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

	r.Get("/plcs", h.PLC.List)
	r.Get("/plcs/{plc_id}", h.PLC.Get)
	r.Get("/plcs/{plc_id}/status", h.PLC.GetStatus)
	r.Get("/plcs/{plc_id}/tags", h.PLC.ListTags)

	r.Get("/tags", h.Tag.List)
	r.Get("/tags/{tag_id}", h.Tag.Get)

	r.Get("/alarms", h.Alarms.List)
	r.Get("/alarms/active", h.Alarms.ListActive)

	r.Get("/collector/status", h.Collector.Status)
	r.Post("/collector/pause", h.Collector.Pause)
	r.Post("/collector/resume", h.Collector.Resume)

	r.Handle("/*", web.Handler())

	return r
}

func routesTelemetryOnly(telemetry *handlers.TelemetryHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get("/telemetry/latest", telemetry.Latest)
	r.Get("/telemetry/history", telemetry.History)

	return r
}
