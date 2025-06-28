package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os/signal"
	"prometheus-net-discovery/internal/config"
	"prometheus-net-discovery/internal/discovery"
	"prometheus-net-discovery/internal/server"
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	listenAddr := flag.String("listenAddr", ":8080", "Listen address of the HTTP server")
	configPath := flag.String("config", "config.yaml", "Path to the config file")
	logLevel := flag.String("logLevel", "info", "Log level")

	flag.Parse()
	log.SetFormatter(&log.JSONFormatter{})

	// set log level
	logLevelParsed, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %s", err)
	}

	log.SetLevel(logLevelParsed)

	// load config
	log.Infof("Load config from %s", *configPath)

	cfg := config.New()
	if err := cfg.Load(*configPath); err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	// describe context
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	g, gCtx := errgroup.WithContext(ctx)

	// run discovery
	d := discovery.New(cfg)

	g.Go(func() error {
		return d.Run(gCtx)
	})

	// run http server
	log.Infof("start http server on %s", *listenAddr)

	srv := server.New(*listenAddr)

	g.Go(func() error {
		return srv.ListenAndServe()
	})

	// graceful shutdown http server
	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), server.ServerTimeoutShutdown)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("server shutdown gracefully")
		} else {
			log.Errorf("error execution discovery: %v", err)
		}
	}
}
