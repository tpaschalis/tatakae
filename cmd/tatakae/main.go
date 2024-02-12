package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	tk "github.com/tpaschalis/tatakae"
)

var (
	defaultEndpoint = "http://localhost:4318"
	defaultHTTPPort = ":2244"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	endpoint, found := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !found {
		endpoint = defaultEndpoint
	}

	metricsEndpoint, mfound := os.LookupEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")
	if !mfound {
		if found {
			metricsEndpoint = endpoint + "/v1/metrics"
		} else {
			metricsEndpoint = defaultEndpoint + "/v1/metrics"
		}
	}

	logsEndpoint, lfound := os.LookupEnv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	if !lfound {
		if found {
			logsEndpoint = endpoint + "/v1/logs"
		} else {
			logsEndpoint = defaultEndpoint + "/v1/logs"
		}
	}

	tracesEndpoint, tfound := os.LookupEnv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	if !tfound {
		if found {
			tracesEndpoint = endpoint + "/v1/traces"
		} else {
			tracesEndpoint = defaultEndpoint + "/v1/traces"
		}
	}

	httpPort, pfound := os.LookupEnv("TATAKAE_PORT")
	if !pfound {
		httpPort = defaultHTTPPort
	}

	cfg := tk.Config{
		MetricsEndpoint: metricsEndpoint,
		LogsEndpoint:    logsEndpoint,
		TracesEndpoint:  tracesEndpoint,
	}

	metricSink, logSink, traceSink, err := tk.NewOTLPHTTPExporter(ctx, logger, cfg)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create the OLTPHTTP exporter", "err", err)
	}
	m := tk.NewMetrics()
	l := tk.NewLogs()
	t := tk.NewTraces()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:              httpPort,
			ReadHeaderTimeout: 5 * time.Second,
		}
		err := server.ListenAndServe()
		if err != nil {
			fmt.Printf("http server exited: %v", err)
			return
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			metricSink.ConsumeMetrics(ctx, m)
			logSink.ConsumeLogs(ctx, l)
			traceSink.ConsumeTraces(ctx, t)
		case <-ctx.Done():
			fmt.Println("Exiting...")
			return
		}
	}
}
