all: airflow-prometheus-exporter
.PHONY: all

airflow-prometheus-exporter: main.go collector.go
	go build .
