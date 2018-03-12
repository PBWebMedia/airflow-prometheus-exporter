package main

import (
	"os"
	"testing"
)

func TestFormulateEventTotalDataPreparedStatementMySQL(t *testing.T) {
	expectedStmt := "SELECT COUNT(*), COALESCE(dag_id, ''), COALESCE(task_id, ''), event, MAX(id) FROM log WHERE id > ? GROUP BY dag_id, task_id, event"
	os.Setenv("AIRFLOW_PROMETHEUS_DATABASE_BACKEND", "mysql")

	returnedStmt := formulateEventTotalDataPreparedStatement()

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}

func TestFormulateEventTotalDataPreparedStatementPostgresSQL(t *testing.T) {
	expectedStmt := "SELECT COUNT(*), COALESCE(dag_id, ''), COALESCE(task_id, ''), event, MAX(id) FROM log WHERE id > $1 GROUP BY dag_id, task_id, event"
	os.Setenv("AIRFLOW_PROMETHEUS_DATABASE_BACKEND", "postgres")

	returnedStmt := formulateEventTotalDataPreparedStatement()

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}
