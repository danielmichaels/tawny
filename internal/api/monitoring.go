package api

import (
	"context"

	"github.com/danielmichaels/tawny/gen/monitoring"
	"github.com/danielmichaels/tawny/internal/logger"
	"github.com/danielmichaels/tawny/internal/version"
)

// monitoring service example implementation.
// The example methods log the requests and return zero values.
type monitoringsrvc struct {
	logger *logger.Logger
}

// NewMonitoring returns the monitoring service implementation.
func NewMonitoring(logger *logger.Logger) monitoring.Service {
	return &monitoringsrvc{logger}
}

// Healthz endpoint
func (s *monitoringsrvc) Healthz(ctx context.Context) (err error) {
	return
}

// Version Application version information endpoint
func (s *monitoringsrvc) Version(ctx context.Context) (res *monitoring.Version2, err error) {
	revision := version.Get()
	if revision == "" {
		return &monitoring.Version2{Version: nil}, nil
	}

	return &monitoring.Version2{Version: &revision}, nil
}
