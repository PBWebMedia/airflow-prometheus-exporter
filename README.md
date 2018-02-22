# Airflow prometheus exporter

[![Docker Hub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/pbweb/airflow-prometheus-exporter/)
[![Docker Pulls](https://img.shields.io/docker/pulls/pbweb/airflow-prometheus-exporter.svg)]()
[![Docker Stars](https://img.shields.io/docker/stars/pbweb/airflow-prometheus-exporter.svg)]()

Export airflow metrics in [Prometheus](https://prometheus.io/) format.

# Build

Requires [Go](https://golang.org/doc/install). Tested with Go 1.9+.

    go get
    go build -o airflow-prometheus-exporter .

# Run

The exporter can be configured using environment variables. These are the defaults:

    AIRFLOW_PROMETHEUS_LISTEN_ADDR=:9112
    AIRFLOW_PROMETHEUS_MYSQL_HOST=localhost
    AIRFLOW_PROMETHEUS_MYSQL_PORT=3306
    AIRFLOW_PROMETHEUS_MYSQL_USER=airflow
    AIRFLOW_PROMETHEUS_MYSQL_PASSWORD=airflow
    AIRFLOW_PROMETHEUS_MYSQL_DATABASE=airflow

Run the exporter:

    ./airflow-prometheus-exporter

The metrics can be scraped from:

    http://localhost:9112/metrics

# Run using docker

Run using docker:

    docker run -p 9112:9112 pbweb/airflow-prometheus-exporter

Or using docker-compose:

    services:
        image: pbweb/airflow-prometheus-exporter
        restart: always
        environment:
            - "AIRFLOW_PROMETHEUS_MYSQL_HOST=mysql.airflow.lan"
        ports:
            - "9112:9112"

# License

See [LICENSE.md](LICENSE.md)
