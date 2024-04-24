package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/danielmichaels/tawny/gen/identity"
	"github.com/danielmichaels/tawny/internal/config"
	"github.com/danielmichaels/tawny/internal/store"
	"github.com/danielmichaels/tawny/internal/webserver"

	identitysvr "github.com/danielmichaels/tawny/gen/http/identity/server"
	monitoringsvr "github.com/danielmichaels/tawny/gen/http/monitoring/server"
	openapisvr "github.com/danielmichaels/tawny/gen/http/openapi/server"
	"github.com/danielmichaels/tawny/gen/monitoring"
	"github.com/danielmichaels/tawny/gen/openapi"
	tawny "github.com/danielmichaels/tawny/internal/api"
	svclogger "github.com/danielmichaels/tawny/internal/logger"
	"github.com/spf13/cobra"
	goahttp "goa.design/goa/v3/http"
	httpmdlwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

func ServeCmd(ctx context.Context) *cobra.Command {
	var isConsole bool
	var debugF bool
	var apiServerOnly bool
	var webServerOnly bool
	cmd := &cobra.Command{
		Use:   "serve",
		Args:  cobra.ExactArgs(0),
		Short: "Runs the servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.AppConfig()
			var (
				logger *svclogger.Logger
			)
			{
				logger = svclogger.New("tawny", debugF, isConsole)
			}
			logger.Info().Int("GOMAXPROCS", runtime.GOMAXPROCS(0)).Send()

			// Initialize stores
			db, err := store.NewDatabasePool(ctx, cfg)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to connect to database")
			}
			defer db.Close()
			err = db.Ping(ctx)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to ping database at startup")
			}
			dbx := store.New(db)
			// Initialise admin user on first boot; this can be updated after the fact
			// todo: find a better way to do this
			err = dbx.BoostrapAdminIfNotExists(ctx, logger)
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to boostrap admin user")
			}

			// Initialize the services.
			var (
				monitoringSvc monitoring.Service
				openapiSvc    openapi.Service
				identitySvc   identity.Service
			)
			{
				monitoringSvc = tawny.NewMonitoring(logger)
				openapiSvc = tawny.NewOpenapi(logger)
				identitySvc = tawny.NewIdentity(logger, dbx)
			}

			// Wrap the services in endpoints that can be invoked from other services
			// potentially running in different processes.
			var (
				monitoringEndpoints *monitoring.Endpoints
				openapiEndpoints    *openapi.Endpoints
				identityEndpoints   *identity.Endpoints
			)
			{
				monitoringEndpoints = monitoring.NewEndpoints(monitoringSvc)
				openapiEndpoints = openapi.NewEndpoints(openapiSvc)
				identityEndpoints = identity.NewEndpoints(identitySvc)
			}

			// Create channel used by both the signal handler and server goroutines
			// to notify the main goroutine when to stop the server.
			errc := make(chan error)

			// Setup interrupt handler. This optional step configures the process so
			// that SIGINT and SIGTERM signals cause the services to stop gracefully.
			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
				errc <- fmt.Errorf("%s", <-c)
			}()

			var wg sync.WaitGroup
			ctx, cancel := context.WithCancel(ctx)

			if apiServerOnly {
				port := cfg.Server.APIPort
				addr := fmt.Sprintf("http://0.0.0.0:%d", port)
				u, err := url.Parse(addr)
				if err != nil {
					logger.Fatal().Msgf("invalid URL %#v: %s\n", addr, err)
				}
				handleHTTPServer(
					ctx,
					u,
					monitoringEndpoints,
					openapiEndpoints,
					identityEndpoints,
					&wg,
					errc,
					logger,
					debugF,
				)
			}
			if webServerOnly {
				app := &webserver.Application{
					Config: cfg,
					Logger: logger,
					DB:     dbx,
				}
				go func() {
					err := app.Serve(ctx)
					if err != nil {
						logger.Error().Err(err).Msg("error in webserver application")
					}
				}()
			}
			if !webServerOnly && !apiServerOnly {
				logger.Fatal().Msgf("must run either API or Web server. nothing selected")
			}
			logger.Info().Msgf("exiting (%v)", <-errc)

			// Send cancellation signal to the goroutines.
			cancel()

			wg.Wait()
			logger.Info().Msg("exited")
			return nil
		},
	}
	cmd.Flags().BoolVar(&debugF, "debug", false, "Log request and response bodies")
	cmd.Flags().BoolVar(&isConsole, "console", false, "Use zerolog ConsoleWriter")
	cmd.Flags().BoolVar(&apiServerOnly, "api", false, "Run the API server")
	cmd.Flags().BoolVar(&webServerOnly, "web", false, "Run the Web server")
	return cmd
}

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(
	ctx context.Context,
	u *url.URL,
	monitoringEndpoints *monitoring.Endpoints,
	openapiEndpoints *openapi.Endpoints,
	identityEndpoints *identity.Endpoints,
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
		identityServer   *identitysvr.Server
	)
	{
		eh := errorHandler(logger)
		monitoringServer = monitoringsvr.New(monitoringEndpoints, mux, dec, enc, eh, nil)
		openapiServer = openapisvr.New(openapiEndpoints, mux, dec, enc, eh, nil)
		identityServer = identitysvr.New(identityEndpoints, mux, dec, enc, eh, nil)
		if debug {
			servers := goahttp.Servers{
				monitoringServer,
				openapiServer,
				identityServer,
			}
			servers.Use(httpmdlwr.Debug(mux, os.Stdout))
		}
	}
	// Configure the mux.
	monitoringsvr.Mount(mux, monitoringServer)
	openapisvr.Mount(mux, openapiServer)
	identitysvr.Mount(mux, identityServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		handler = httpmdlwr.Log(adapter)(handler) //nolint:all
		handler = httpmdlwr.RequestID()(handler)  //nolint:all
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
	for _, m := range identityServer.Mounts {
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
