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
	loadEnv()

	c := newCollector(dbDsn)
	prometheus.Register(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadEnv() {
	mysqlHost := getEnvOr("AIRFLOW_PROMETHEUS_MYSQL_HOST", "localhost")
	mysqlPort := getEnvOr("AIRFLOW_PROMETHEUS_MYSQL_PORT", "3306")
	mysqlUser := getEnvOr("AIRFLOW_PROMETHEUS_MYSQL_USER", "airflow")
	mysqlPassword := getEnvOr("AIRFLOW_PROMETHEUS_MYSQL_PASSWORD", "airflow")
	mysqlDb := getEnvOr("AIRFLOW_PROMETHEUS_MYSQL_DATABASE", "airflow")

	addr = getEnvOr("AIRFLOW_PROMETHEUS_LISTEN_ADDR", ":9112")
	dbDsn = mysqlUser + ":" + mysqlPassword + "@(" + mysqlHost + ":" + mysqlPort + ")/" + mysqlDb
}

func getEnvOr(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
