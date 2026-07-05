package api

import (
	"github.com/go-chi/chi/v5"

	"pharma-platform/internal/api/handlers"
)

func routes(
	telemetry *handlers.TelemetryHandler,
) *chi.Mux {

	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get(
		"/telemetry/latest",
		telemetry.Latest,
	)

	r.Get(
		"/telemetry/history",
		telemetry.History,
	)

	return r
}
