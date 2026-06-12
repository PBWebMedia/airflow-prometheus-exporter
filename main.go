package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	driverMySQL    = "mysql"
	driverPostgres = "postgres"
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
	if err := prometheus.Register(c); err != nil {
		log.Fatal("Failed to register collector: ", err)
	}

	log.Print("Listening on: ", addr)
	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func loadEnv() {
	databaseDefaultPort := map[string]string{
		driverMySQL:    "3306",
		driverPostgres: "5432",
	}

	databaseBackend := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_BACKEND", driverMySQL)
	databaseHost := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_HOST", "localhost")
	databasePort := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PORT", databaseDefaultPort[databaseBackend])

	if databaseBackend != driverMySQL && databaseBackend != driverPostgres {
		log.Fatal("airflow-exporter: Unknown database backend specified: ", databaseBackend)
	}

	databaseUser := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_USER", "airflow")
	databasePassword := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_PASSWORD", "airflow")
	databaseName := getEnvOr("AIRFLOW_PROMETHEUS_DATABASE_NAME", "airflow")

	addr = getEnvOr("AIRFLOW_PROMETHEUS_LISTEN_ADDR", ":9112")
	dbDriver = databaseBackend

	switch databaseBackend {
	case driverMySQL:
		dbDsn = databaseUser + ":" + databasePassword + "@(" + databaseHost + ":" + databasePort + ")/" + databaseName
		dbDsnPasswordMasked = databaseUser + ":********@(" + databaseHost + ":" + databasePort + ")/" + databaseName
	case driverPostgres:
		properties := map[string]string{
			"user":        databaseUser,
			"password":    databasePassword,
			"host":        databaseHost,
			"port":        databasePort,
			"dbname":      databaseName,
			"sslmode":     getEnvOr("AIRFLOW_PROMETHEUS_POSTGRES_SSL_MODE", "disable"),
			"sslcert":     getEnvOr("AIRFLOW_PROMETHEUS_POSTGRES_SSL_CERT", ""),
			"sslkey":      getEnvOr("AIRFLOW_PROMETHEUS_POSTGRES_SSL_KEY", ""),
			"sslrootcert": getEnvOr("AIRFLOW_PROMETHEUS_POSTGRES_SSL_ROOT_CERT", ""),
		}

		dbDsn = createPostgresDsn(properties)

		if properties["password"] != "" {
			properties["password"] = "********"
		}
		dbDsnPasswordMasked = createPostgresDsn(properties)
	}
}

func getEnvOr(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func createPostgresDsn(properties map[string]string) string {
	list := make([]string, 0)
	for key, value := range properties {
		if value != "" {
			list = append(list, key+"="+value)
		}
	}

	return strings.Join(list, " ")
}
