all: airflow-prometheus-exporter
.PHONY: all

airflow-prometheus-exporter: main.go collector.go
	CGO_ENABLED=0 go build .
