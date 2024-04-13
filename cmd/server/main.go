package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	monitoring "github.com/danielmichaels/lappycloud/gen/monitoring"
	openapi "github.com/danielmichaels/lappycloud/gen/openapi"
	lappycloud "github.com/danielmichaels/lappycloud/internal/api"
	svclogger "github.com/danielmichaels/lappycloud/internal/logger"
)

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		hostF   = flag.String("host", "localhost", "Server host (valid values: localhost)")
		domainF = flag.String(
			"domain",
			"",
			"Host domain name (overrides host domain specified in service design)",
		)
		httpPortF = flag.String(
			"http-port",
			"",
			"HTTP port (overrides host HTTP port specified in service design)",
		)
		secureF   = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")
		isConsole = flag.Bool("console", false, "Use zerolog ConsoleWriter")
	)
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	var (
		logger *svclogger.Logger
	)
	{
		logger = svclogger.New("api", *dbgF, *isConsole)
	}

	// Initialize the services.
	var (
		monitoringSvc monitoring.Service
		openapiSvc    openapi.Service
	)
	{
		monitoringSvc = lappycloud.NewMonitoring(logger)
		openapiSvc = lappycloud.NewOpenapi(logger)
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var (
		monitoringEndpoints *monitoring.Endpoints
		openapiEndpoints    *openapi.Endpoints
	)
	{
		monitoringEndpoints = monitoring.NewEndpoints(monitoringSvc)
		openapiEndpoints = openapi.NewEndpoints(openapiSvc)
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
	ctx, cancel := context.WithCancel(context.Background())

	// Start the servers and send errors (if any) to the error channel.
	switch *hostF {
	case "localhost":
		{
			addr := "http://localhost:9090"
			u, err := url.Parse(addr)
			if err != nil {
				logger.Fatal().Msgf("invalid URL %#v: %s\n", addr, err)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					logger.Fatal().Msgf("invalid URL %#v: %s\n", addr, err)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(
				ctx,
				u,
				monitoringEndpoints,
				openapiEndpoints,
				&wg,
				errc,
				logger,
				*dbgF,
			)
		}

	default:
		logger.Fatal().Msgf("invalid host argument: %q (valid hosts: localhost)\n", *hostF)
	}

	// Wait for signal.
	logger.Info().Msgf("exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	logger.Info().Msg("exited")
}
