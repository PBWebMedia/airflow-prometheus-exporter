package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
)

var (
	addr  string
	dbDsn string
)

func main() {
	log.Print("Starting airflow-exporter")
	loadEnv()

	log.Print("Connecting to: ", dbDsn)
	c := newCollector(dbDsn)
	prometheus.Register(c)

	log.Print("Listening on: ", addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadEnv() {
	databaseBackend := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_BACKEND", "mysql")
	databaseHost := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_HOST", "localhost")
	databasePort := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PORT", "3306")

	if databaseBackend == "mysql" {
		databasePort = getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PORT", "3306")
	} else if databaseBackend == "postgresql" {
		databasePort = getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PORT", "5432")
	} else {
		log.Fatal("airflow-exporter: Unknown database backend specified: ", databaseBackend)
	}

	databaseUser := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_USER", "airflow")
	databasePassword := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PASSWORD", "airflow")
	databaseName := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_NAME", "airflow")

	addr = getEnvOr("AIRFLOW_PROMETHEUS_LISTEN_ADDR", ":9112")
	dbDsn = databaseBackend + "://" + databaseUser + ":" + databasePassword + "@(" + databaseHost + ":" + databasePort + ")/" + databaseName
}

func getEnvOr(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
