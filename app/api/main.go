package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Yeremi528/laboratorio/business/web/debug"
	"github.com/Yeremi528/laboratorio/foundation/logger"
	"github.com/Yeremi528/laboratorio/foundation/web"
	"github.com/ardanlabs/conf/v3"
)

var build = "dev"

// @title Laboratorio Dev
// @version 1.0
// @description This is documentation for Laboratorio Dev API
// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /lab

func main() {

	var logLevel logger.Level
	level := os.Getenv("APP_LOG_LEVEL")
	switch level {
	case "INFO":
		logLevel = logger.LevelInfo
	case "DEBUG":
		logLevel = logger.LevelDebug
	case "WARN":
		logLevel = logger.LevelWarn
	case "ERROR":
		logLevel = logger.LevelError
	default:
		level = "INFO"
		logLevel = logger.LevelInfo
	}

	traceFunc := func(ctx context.Context) []any {
		v := web.GetValues(ctx)

		fields := make([]any, 2, 4)
		fields[0], fields[1] = "traceID", v.TraceID

		return fields
	}

	log := logger.New(os.Stdout, logLevel, "go-ms-laboratorio", traceFunc)

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout        time.Duration `conf:"default:5s"`
			WriteTimeout       time.Duration `conf:"default:10s"`
			IdleTimeout        time.Duration `conf:"default:120s"`
			ShutdownTimeout    time.Duration `conf:"default:20s"`
			APIHost            string        `conf:"default:0.0.0.0:3000"`
			DebugHost          string        `conf:"default:0.0.0.0:4000"`
			CORSAllowedOrigins []string      `conf:"default:*"`
		}
		Auth struct {
			KeysFolder string `conf:"default:zarf/keys/"`
			ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			Issuer     string `conf:"default:service project"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			HostPort     string `conf:"default:database-service.sales-system.svc.cluster.local"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Tempo struct {
			ReporterURI string  `conf:"default:tempo.sales-system.svc.cluster.local:4317"`
			ServiceName string  `conf:"default:sales-api"`
			Probability float64 `conf:"default:0.05"`
			// Shouldn't use a high Probability value in non-developer systems.
			// 0.05 should be enough for most systems. Some might want to have
			// this even lower.
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Service Project",
		},
	}

	const prefix = "APP"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	expvar.NewString("build").Set(build)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "hostport", cfg.DB.HostPort)

	db, err := pgx.Open(pgx.Config{
		User:            hiddenAppConfig.Postgres.User,
		Password:        hiddenAppConfig.Postgres.Password,
		Host:            hiddenAppConfig.Postgres.Host,
		Port:            hiddenAppConfig.Postgres.Port,
		Name:            hiddenAppConfig.Postgres.Name,
		MaxIdleConns:    hiddenAppConfig.Postgres.MaxIdleConns,
		MaxOpenConns:    hiddenAppConfig.Postgres.MaxOpenConns,
		IdleConnTimeout: hiddenAppConfig.Postgres.ConnMaxIdleTime,
		EnableTLS:       hiddenAppConfig.Postgres.EnableTLS,
		CACert:          tmpServerCA,
		ClientCert:      tmpClientCert,
		ClientKey:       tmpClientKey,
		ApplicationName: "onboarding/go-ms-enrollment-finalize",
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping database support", "hostport", cfg.DB.HostPort)
		db.Close()
	}()

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      debug.Mux(),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil

}
