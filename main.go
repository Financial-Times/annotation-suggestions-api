package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/Financial-Times/api-endpoint"
	"github.com/Financial-Times/go-ft-http/fthttp"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"

	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/health"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

const (
	appDescription = "Provides suggestions for draft content."
	defaultAppName = "draft-content-suggestions"
)

func main() {
	app := cli.App(defaultAppName, appDescription)

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "draft-content-suggestions",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})
	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  defaultAppName,
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the OpenAPI YML file.",
		EnvVar: "API_YML",
	})
	draftContentEndpoint := app.String(cli.StringOpt{
		Name:   "draft-content-endpoint",
		Value:  "http://draft-content-public-read:8080/content",
		Desc:   "Endpoint for Draft Content API",
		EnvVar: "DRAFT_CONTENT_ENDPOINT",
	})
	draftContentGtgEndpoint := app.String(cli.StringOpt{
		Name:   "draft-content-gtg-endpoint",
		Value:  "http://draft-content-public-read:8080/__gtg",
		Desc:   "GTG Endpoint for Draft Content API",
		EnvVar: "DRAFT_CONTENT_GTG_ENDPOINT",
	})
	suggestionsEndpoint := app.String(cli.StringOpt{
		Name:   "suggestions-umbrella-endpoint",
		Value:  "http://test.api.ft.com/content/suggest",
		Desc:   "Endpoint for Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_ENDPOINT",
	})
	suggestionsGtgEndpoint := app.String(cli.StringOpt{
		Name:   "suggestions-umbrella-gtg-endpoint",
		Value:  "http://test.api.ft.com/content/suggest/__gtg",
		Desc:   "Endpoint for Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_GTG_ENDPOINT",
	})
	suggestionsAPIKey := app.String(cli.StringOpt{
		Name:   "suggestions-api-key",
		Value:  "",
		Desc:   "API key to access Suggestions Umbrella",
		EnvVar: "SUGGESTIONS_API_KEY",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Value:  "info",
		Desc:   "Logging Level",
		EnvVar: "LOG_LEVEL",
	})

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Warnf("Log level %s could not be parsed, defaulting to info", *logLevel)
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&log.JSONFormatter{})

	app.Action = func() {
		log.Infof("[Startup] %s is starting", *appName)
		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		// We don't want logging for GTG requests in the middleware
		healthCl := fthttp.NewClientBuilder().
			WithTimeout(10*time.Second).
			WithSysInfo("PAC", *appSystemCode).
			Build()
		loggingCl := fthttp.NewClientBuilder().
			WithTimeout(10*time.Second).
			WithSysInfo("PAC", *appSystemCode).
			WithLogging(log.StandardLogger()).
			Build()
		contentAPI, tmpErr := draft.NewContentAPI(*draftContentEndpoint, *draftContentGtgEndpoint, loggingCl, healthCl)
		if tmpErr != nil {
			log.WithError(tmpErr).Error("Draft Content API error, exiting ...")
			return
		}

		umbrellaAPI, tmpErr := suggestions.NewUmbrellaAPI(*suggestionsEndpoint, *suggestionsGtgEndpoint, *suggestionsAPIKey, loggingCl, healthCl)
		if tmpErr != nil {
			log.WithError(tmpErr).Error("Suggestions Umbrella API error, exiting ...")
			return
		}

		serveEndpoints(*appSystemCode, *appName, *port, apiYml, requestHandler{contentAPI, umbrellaAPI})
	}

	err = app.Run(os.Args)
	if err != nil {
		log.WithError(err).Errorf("%s could not start!", defaultAppName)
		return
	}
}

func serveEndpoints(appSystemCode string, appName string, port string, apiYml *string, requestHandler requestHandler) {
	healthService := health.NewHealthService(appSystemCode, appName, appDescription,
		requestHandler.dca, requestHandler.sua)

	serveMux := http.NewServeMux()

	serveMux.HandleFunc(health.DefaultHealthPath, http.HandlerFunc(fthealth.Handler(healthService.Health())))
	serveMux.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.GTG))
	serveMux.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	if apiYml != nil {
		apiEndpoint, err := api.NewAPIEndpointForFile(*apiYml)
		if err != nil {
			log.WithError(err).WithField("file", apiYml).Warn("Failed to serve the API Endpoint for this service. Please validate the file exists, and that it fits the OpenAPI specification.")
		} else {
			serveMux.HandleFunc(api.DefaultPath, apiEndpoint.ServeHTTP)
		}
	}

	servicesRouter := mux.NewRouter()
	servicesRouter.HandleFunc("/drafts/content/{uuid}/suggestions",
		requestHandler.draftContentSuggestionsRequest).Methods("GET")

	monitoringRouter := httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), servicesRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	serveMux.Handle("/", monitoringRouter)

	server := &http.Server{Addr: ":" + port, Handler: serveMux}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.WithError(err).Info("HTTP server closing with message")
		}
		wg.Done()
	}()

	waitForSignal()
	log.Infof("[Shutdown] %s is shutting down", defaultAppName)

	if err := server.Close(); err != nil {
		log.WithError(err).Error("Unable to stop http server")
	}

	wg.Wait()
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
