package main

import (
	"context"
	svclogger "github.com/danielmichaels/lappycloud/internal/logger"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	monitoringsvr "github.com/danielmichaels/lappycloud/gen/http/monitoring/server"
	openapisvr "github.com/danielmichaels/lappycloud/gen/http/openapi/server"
	monitoring "github.com/danielmichaels/lappycloud/gen/monitoring"
	openapi "github.com/danielmichaels/lappycloud/gen/openapi"
	goahttp "goa.design/goa/v3/http"
	httpmdlwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(
	ctx context.Context,
	u *url.URL,
	monitoringEndpoints *monitoring.Endpoints,
	openapiEndpoints *openapi.Endpoints,
	wg *sync.WaitGroup,
	errc chan error,
	logger *svclogger.Logger,
	debug bool,
) {

	// Setup goa log adapter.
	var (
		adapter middleware.Logger
	)
	{
		adapter = logger
	}

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/implement/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
		monitoringServer *monitoringsvr.Server
		openapiServer    *openapisvr.Server
	)
	{
		eh := errorHandler(logger)
		monitoringServer = monitoringsvr.New(monitoringEndpoints, mux, dec, enc, eh, nil)
		openapiServer = openapisvr.New(openapiEndpoints, mux, dec, enc, eh, nil)
		if debug {
			servers := goahttp.Servers{
				monitoringServer,
				openapiServer,
			}
			servers.Use(httpmdlwr.Debug(mux, os.Stdout))
		}
	}
	// Configure the mux.
	monitoringsvr.Mount(mux, monitoringServer)
	openapisvr.Mount(mux, openapiServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		handler = httpmdlwr.Log(adapter)(handler)
		handler = httpmdlwr.RequestID()(handler)
	}

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler, ReadHeaderTimeout: time.Second * 60}
	for _, m := range monitoringServer.Mounts {
		logger.Debug().Msgf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range openapiServer.Mounts {
		logger.Debug().Msgf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start HTTP server in a separate goroutine.
		go func() {
			logger.Info().Msgf("API server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		logger.Info().Msgf("shutting down API server at %q", u.Host)

		// Shutdown gracefully with a 30s timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			logger.Fatal().Msgf("failed to shutdown: %v", err)
		}
	}()
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *svclogger.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		_, _ = w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}
