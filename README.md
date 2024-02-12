# tatakae

Tatakae is a data generation tool for testing OpenTelemetry consumers such as
the OpenTelemetry Collector and Grafana Agent Flow.

It generates metrics, logs and traces data and sends them over to a remote
endpoint every 5 seconds using the OTLP HTTP exporter.

The remote endpoint can be adjusted using the following env variables
* `OTEL_EXPORTER_OTLP_ENDPOINT`: sets the endpoint for all three signals. Defaults to `http://localhost:4318`
* `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`: sets the metrics endpoint. Defaults to `${OTEL_EXPORTER_OTLP_ENDPOINT}/v1/metrics`
* `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`: sets the logs endpoint. Defaults to `${OTEL_EXPORTER_OTLP_ENDPOINT}/v1/logs`
* `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: sets the traces endpoint. Defaults to `${OTEL_EXPORTER_OTLP_ENDPOINT}/v1/traces`

Tatakae also serves its own `/metrics` endpoint. (Although at the time no useful metrics have been wired in).

## Build and run go binary

```
$ make build
$ ./build/tatakae
level=debug msg="Preparing to make HTTP request" url=http://localhost:4318/v1/traces
level=debug msg="Preparing to make HTTP request" url=http://localhost:4318/v1/metrics
level=debug msg="Preparing to make HTTP request" url=http://localhost:4318/v1/logs
```

## TODO
* [ ] Make it easier to run as a library
* [ ] Make more parameters configurable (eg. rate of data being sent)
* [ ] Make data generation configurable

## What's with the name?

Means "fight!" in japanese. I just done binging Attack on Titan, and the stupid
protagonist was saying this word every 10 minutes.
