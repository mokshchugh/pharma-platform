package api

import (
	"context"
	"fmt"
	"net/http"

	"pharma-platform/internal/api/handlers"
	"pharma-platform/internal/config"
)

type Server struct {
	http *http.Server
}

func NewFull(cfg config.APIConfig, h *Handlers) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return &Server{
		http: &http.Server{
			Addr:    addr,
			Handler: Routes(h),
		},
	}
}

func NewTelemetryOnly(cfg config.APIConfig, telemetry *handlers.TelemetryHandler) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return &Server{
		http: &http.Server{
			Addr:    addr,
			Handler: routesTelemetryOnly(telemetry),
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
