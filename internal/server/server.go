package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	ServerReadHeaderTimeout = 10 * time.Second
	ServerTimeoutShutdown   = 10 * time.Second
	MetricsSlowTime         = 10
)

func New(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: ServerReadHeaderTimeout,
	}
}
