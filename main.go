package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
)

var (
	addr                string
	dbDriver            string
	dbDsn               string
	dbDsnPasswordMasked string
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Print("Starting airflow-exporter")
	loadEnv()

	log.Print("Connecting to: ", dbDriver, "://", dbDsnPasswordMasked)
	c := newCollector(dbDriver, dbDsn)
	prometheus.Register(c)

	log.Print("Listening on: ", addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadEnv() {
	databaseDefaultPort := map[string]string{
		"mysql":    "3306",
		"postgres": "5432",
	}

	databaseBackend := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_BACKEND", "mysql")
	databaseHost := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_HOST", "localhost")
	databasePort := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PORT", databaseDefaultPort[databaseBackend])

	if !(databaseBackend == "mysql" || databaseBackend == "postgres") {
		log.Fatal("airflow-exporter: Unknown database backend specified: ", databaseBackend)
	}

	databaseUser := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_USER", "airflow")
	databasePassword := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PASSWORD", "airflow")
	databaseName := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_NAME", "airflow")

	addr = getEnvOr("AIRFLOW_PROMETHEUS_LISTEN_ADDR", ":9112")
	dbDriver = databaseBackend

	if databaseBackend == "mysql" {
		dbDsn = databaseUser + ":" + databasePassword + "@(" + databaseHost + ":" + databasePort + ")/" + databaseName
		dbDsnPasswordMasked = databaseUser + ":********@(" + databaseHost + ":" + databasePort + ")/" + databaseName
	} else if databaseBackend == "postgres" {
		dbDsn = "user=" + databaseUser + " password=" + databasePassword + " host=" + databaseHost + " port=" + databasePort + " dbname=" + databaseName + " sslmode=disable"
		dbDsnPasswordMasked = "user=" + databaseUser + " password=******** host=" + databaseHost + " port=" + databasePort + " dbname=" + databaseName + " sslmode=disable"
	}
}

func getEnvOr(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
