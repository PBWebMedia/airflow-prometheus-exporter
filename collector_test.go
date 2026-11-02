package main

import (
	"testing"
)

func TestFormulateEventTotalDataPreparedStatementMySQL(t *testing.T) {
	expectedStmt := "SELECT COUNT(*), COALESCE(dag_id, ''), COALESCE(task_id, ''), event, MAX(id) FROM log WHERE id > ? GROUP BY dag_id, task_id, event"
	dbDriver = driverMySQL

	returnedStmt := formulateEventTotalDataPreparedStatement()

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}

func TestFormulateDagStaleColumnStatementMySQL(t *testing.T) {
	expectedStmt := "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'dag' AND column_name = 'is_stale'"
	dbDriver = driverMySQL

	returnedStmt := formulateDagStaleColumnStatement()

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}

func TestFormulateDagStaleColumnStatementPostgresSQL(t *testing.T) {
	expectedStmt := "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'dag' AND column_name = 'is_stale'"
	dbDriver = driverPostgres

	returnedStmt := formulateDagStaleColumnStatement()

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}

func TestFormulateDagStatementAirflow2(t *testing.T) {
	expectedStmt := "SELECT dag_id, is_paused, is_active FROM dag"

	returnedStmt := formulateDagStatement(false)

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}

func TestFormulateDagStatementAirflow3(t *testing.T) {
	expectedStmt := "SELECT dag_id, is_paused, NOT is_stale FROM dag"

	returnedStmt := formulateDagStatement(true)

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}

func TestFormulateEventTotalDataPreparedStatementPostgresSQL(t *testing.T) {
	expectedStmt := "SELECT COUNT(*), COALESCE(dag_id, ''), COALESCE(task_id, ''), event, MAX(id) FROM log WHERE id > $1 GROUP BY dag_id, task_id, event"
	dbDriver = driverPostgres

	returnedStmt := formulateEventTotalDataPreparedStatement()

	if returnedStmt != expectedStmt {
		t.Errorf("Returned SQL statement was incorrect, got: '%s', want: '%s'.", returnedStmt, expectedStmt)
	}
}
