package tatakae

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util/zapadapter"
	"github.com/prometheus/client_golang/prometheus"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelexporter "go.opentelemetry.io/collector/exporter"
	sdkprometheus "go.opentelemetry.io/otel/exporters/prometheus"
)

var (
	defaultTraceID = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	defaultSpanID  = [8]byte{101, 102, 103, 104, 105, 106, 107, 108}

	DefaultEndpoint = "http://localhost:4318"
	DefaultConfig   = Config{
		MetricsEndpoint: DefaultEndpoint + "/v1/metrics",
		LogsEndpoint:    DefaultEndpoint + "/v1/logs",
		TracesEndpoint:  DefaultEndpoint + "/v1/traces",
	}
)

type Config struct {
	MetricsEndpoint string
	LogsEndpoint    string
	TracesEndpoint  string
}

func NewOTLPHTTPExporter(ctx context.Context, logger log.Logger, cfg Config) (otelexporter.Metrics, otelexporter.Logs, otelexporter.Traces, error) {
	factory := otlphttpexporter.NewFactory()

	// TODO(@tpaschalis) Use to instrument Prometheus metrics.
	reg := prometheus.NewRegistry()
	promExporter, err := sdkprometheus.New(sdkprometheus.WithRegisterer(reg), sdkprometheus.WithoutTargetInfo())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	metricOpts := []metric.Option{metric.WithReader(promExporter)}
	settings := otelexporter.CreateSettings{
		TelemetrySettings: otelcomponent.TelemetrySettings{
			Logger: zapadapter.New(logger),

			TracerProvider: noop.NewTracerProvider(),
			MeterProvider:  metric.NewMeterProvider(metricOpts...),

			ReportComponentStatus: func(*otelcomponent.StatusEvent) error {
				return nil
			},
		},

		BuildInfo: otelcomponent.BuildInfo{
			Command:     "tatakae",
			Description: "a simple data generator for OpenTelemetry",
			Version:     "v0.0.1",
		},
	}

	exporterConfig := &otlphttpexporter.Config{
		HTTPClientSettings: confighttp.NewDefaultHTTPClientSettings(),
		QueueSettings:      exporterhelper.NewDefaultQueueSettings(),
		RetrySettings:      exporterhelper.NewDefaultRetrySettings(),
		MetricsEndpoint:    cfg.MetricsEndpoint,
		LogsEndpoint:       cfg.LogsEndpoint,
		TracesEndpoint:     cfg.TracesEndpoint,
	}

	var metricsExporter otelexporter.Metrics
	var logsExporter otelexporter.Logs
	var tracesExporter otelexporter.Traces

	metricsExporter, err = factory.CreateMetricsExporter(ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return nil, nil, nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	logsExporter, err = factory.CreateLogsExporter(ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return nil, nil, nil, fmt.Errorf("failed to create logs exporter: %w", err)
	}
	tracesExporter, err = factory.CreateTracesExporter(ctx, settings, exporterConfig)
	if err != nil && !errors.Is(err, otelcomponent.ErrDataTypeIsNotSupported) {
		return nil, nil, nil, fmt.Errorf("failed to create traces exporter: %w", err)
	}
	host := NewHost(logger)
	metricsExporter.Start(ctx, host)
	logsExporter.Start(ctx, host)
	tracesExporter.Start(ctx, host)

	return metricsExporter, logsExporter, tracesExporter, nil
}

func NewMetrics() pmetric.Metrics {
	ts := pcommon.NewTimestampFromTime(time.Now())
	m := pmetric.NewMetrics()

	rm := m.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("resource-foo", "resource-bar")
	rm.SetSchemaUrl("https://example.com")

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.SetSchemaUrl("https://example2.com")
	sm.Scope().SetName("scope-name")
	sm.Scope().SetVersion("scope-version")
	sm.Scope().SetDroppedAttributesCount(0)
	sm.Scope().Attributes().PutStr("scope-foo", "scope-bar")

	mr := sm.Metrics().AppendEmpty()
	mr.SetName("metric-name")
	mr.SetDescription("a description of the metric")
	mr.SetUnit("second")

	g := mr.SetEmptyGauge()
	dp := g.DataPoints().AppendEmpty()

	dp.SetFlags(pmetric.DefaultDataPointFlags)
	dp.SetTimestamp(ts)
	dp.SetStartTimestamp(ts)
	dp.SetIntValue(147)
	dp.Attributes().PutStr("metric-foo", "metric-bar")

	e := dp.Exemplars().AppendEmpty()
	e.SetTimestamp(ts)
	e.SetTraceID(defaultTraceID)
	e.SetSpanID(defaultSpanID)
	e.SetIntValue(42)
	e.FilteredAttributes().PutStr("exemplar-foo", "exemplar-bar")

	return m
}

func NewLogs() plog.Logs {
	ts := pcommon.NewTimestampFromTime(time.Now())
	l := plog.NewLogs()

	rl := l.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("resource-foo", "resource-bar")
	rl.SetSchemaUrl("https://example.com")

	sl := rl.ScopeLogs().AppendEmpty()
	sl.SetSchemaUrl("https://example2.com")
	sl.Scope().SetName("scope-name")
	sl.Scope().SetVersion("scope-version")
	sl.Scope().SetDroppedAttributesCount(0)
	sl.Scope().Attributes().PutStr("scope-foo", "scope-bar")

	lr := sl.LogRecords().AppendEmpty()
	lr.Attributes().PutStr("record-foo", "record-bar")
	lr.SetDroppedAttributesCount(0)
	lr.SetObservedTimestamp(ts)
	lr.SetTimestamp(ts)
	lr.SetSeverityText("warn")
	lr.SetSeverityNumber(plog.SeverityNumberWarn)
	lr.SetSpanID(defaultSpanID)
	lr.SetTraceID(defaultTraceID)
	lr.SetFlags(plog.DefaultLogRecordFlags)

	lr.Body().SetStr("a warn log message")

	return l
}

func NewTraces() ptrace.Traces {
	ts := pcommon.NewTimestampFromTime(time.Now())
	t := ptrace.NewTraces()

	rs := t.ResourceSpans().AppendEmpty()
	rs.SetSchemaUrl("https://example.com")
	rs.Resource().Attributes().PutStr("resource-foo", "resource-bar")

	ss := rs.ScopeSpans().AppendEmpty()
	ss.SetSchemaUrl("https://example2.com")
	ss.Scope().SetDroppedAttributesCount(0)
	ss.Scope().SetName("scope-name")
	ss.Scope().SetVersion("scope-version")
	ss.Scope().Attributes().PutStr("scope-foo", "scope-bar")

	s := ss.Spans().AppendEmpty()
	s.SetTraceID(defaultTraceID)
	s.SetSpanID(defaultSpanID)
	// No Setter for TraceState; maybe MoveTo?
	s.SetName("span-name")
	s.SetStartTimestamp(ts)
	s.SetEndTimestamp(ts)
	s.SetKind(ptrace.SpanKindServer)
	s.SetDroppedAttributesCount(0)
	s.SetDroppedEventsCount(0)
	s.SetDroppedLinksCount(0)
	s.SetParentSpanID(pcommon.NewSpanIDEmpty())
	// No Setter for status
	s.Attributes().PutStr("span-foo", "span-bar")

	l := s.Links().AppendEmpty()
	l.SetTraceID(defaultTraceID)
	l.SetSpanID(defaultSpanID)
	// No Setter for TraceState; maybe MoveTo?
	l.SetDroppedAttributesCount(0)
	l.Attributes().PutStr("link-foo", "link-bar")

	e := s.Events().AppendEmpty()
	e.SetName("event-name")
	e.SetTimestamp(ts)
	e.SetDroppedAttributesCount(0)
	e.Attributes().PutStr("event-foo", "event-bar")

	return t
}
