package main

import (
	"database/sql"
	"log"
	"math"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

type collector struct {
	dbDriver        string
	dbDsn           string
	eventTotalCache eventTotalCache
	up              *prometheus.Desc
	dagActive       *prometheus.Desc
	dagPaused       *prometheus.Desc
	eventTotal      *prometheus.Desc
	scrapeFailures  *prometheus.Desc
	dagRunStates    *prometheus.Desc
	poolSlots       *prometheus.Desc
	failureCount    int
}

type eventTotalCache struct {
	mutex       *sync.Mutex
	data        map[string]map[string]map[string]float64
	lastEventID float64
}

type dag struct {
	dag    string
	paused bool
	subdag bool
	active bool
}

type eventTotal struct {
	count float64
	dag   string
	task  string
	event string
}

type dagRunState struct {
	count float64
	dag   string
	state string
}

type poolSlot struct {
	name string
	size float64
}

type metrics struct {
	dagList      []dag
	eventTotals  []eventTotal
	dagRunStates []dagRunState
	poolSlots    []poolSlot
}

const metricsNamespace = "airflow"

var boolToFloat64 = map[bool]float64{true: 1.0, false: 0.0}

func newFuncMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(metricsNamespace, "", metricName), docString, labels, nil)
}

func newCollector(dbDriver string, dbDsn string) *collector {
	return &collector{
		dbDriver: dbDriver,
		dbDsn:    dbDsn,
		eventTotalCache: eventTotalCache{
			mutex: &sync.Mutex{},
			data:  make(map[string]map[string]map[string]float64),
		},
		up:             newFuncMetric("up", "able to contact airflow database", nil),
		dagActive:      newFuncMetric("dag_active", "Is the DAG active?", []string{"dag"}),
		dagPaused:      newFuncMetric("dag_paused", "Is the DAG paused?", []string{"dag"}),
		eventTotal:     newFuncMetric("event_total", "Total events per DAG, task and event type", []string{"dag", "task", "event"}),
		scrapeFailures: newFuncMetric("scrape_failures_total", "Number of errors while scraping airflow database", nil),
		dagRunStates:   newFuncMetric("dag_run_state", "Number of DAG runs per DAG and state", []string{"dag", "state"}),
		poolSlots:      newFuncMetric("pool_slots", "Pool name with slot size", []string{"name"}),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.dagActive
	ch <- c.dagPaused
	ch <- c.eventTotal
	ch <- c.scrapeFailures
	ch <- c.dagRunStates
	ch <- c.poolSlots
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	up := 1.0

	m, err := getData(c)
	if err != nil {
		up = 0.0
		c.failureCount++

		log.Println("Error while collecting data from database: " + err.Error())
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)
	ch <- prometheus.MustNewConstMetric(c.scrapeFailures, prometheus.CounterValue, float64(c.failureCount))

	if up == 0.0 {
		return
	}

	for _, dag := range m.dagList {
		ch <- prometheus.MustNewConstMetric(c.dagActive, prometheus.GaugeValue, boolToFloat64[dag.active], dag.dag)
		ch <- prometheus.MustNewConstMetric(c.dagPaused, prometheus.GaugeValue, boolToFloat64[dag.paused], dag.dag)
	}

	for _, et := range m.eventTotals {
		ch <- prometheus.MustNewConstMetric(c.eventTotal, prometheus.CounterValue, et.count, et.dag, et.task, et.event)
	}

	for _, st := range m.dagRunStates {
		ch <- prometheus.MustNewConstMetric(c.dagRunStates, prometheus.GaugeValue, st.count, st.dag, st.state)
	}

	for _, st := range m.poolSlots {
		ch <- prometheus.MustNewConstMetric(c.poolSlots, prometheus.GaugeValue, st.size, st.name)
	}

	return
}

func getData(c *collector) (metrics, error) {
	var m metrics

	db, err := sql.Open(c.dbDriver, c.dbDsn)
	if err != nil {
		return m, err
	}
	defer db.Close()

	m.dagList, err = getDagData(db)
	if err != nil {
		return m, err
	}

	m.eventTotals, err = getEventTotalData(&c.eventTotalCache, db)
	if err != nil {
		return m, err
	}

	m.dagRunStates, err = getDagRunStateData(db)
	if err != nil {
		return m, err
	}

	m.poolSlots, err = getPoolSlotData(db)
	if err != nil {
		return m, err
	}

	return m, err
}

func getDagData(db *sql.DB) ([]dag, error) {
	rows, err := db.Query("SELECT dag_id, is_paused, is_subdag, is_active FROM dag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dagList []dag
	for rows.Next() {
		var dag dag

		err := rows.Scan(&dag.dag, &dag.paused, &dag.subdag, &dag.active)
		if err != nil {
			return nil, err
		}

		dagList = append(dagList, dag)
	}

	return dagList, nil
}

func formulateEventTotalDataPreparedStatement() string {
	databasePreparedSyntax := map[string]string{
		"mysql":    "WHERE id > ?",
		"postgres": "WHERE id > $1",
	}

	preparedStmt := "SELECT COUNT(*), COALESCE(dag_id, ''), COALESCE(task_id, ''), event, MAX(id) FROM log "
	preparedStmt += databasePreparedSyntax[dbDriver]
	preparedStmt += " GROUP BY dag_id, task_id, event"

	return preparedStmt
}

func getEventTotalData(c *eventTotalCache, db *sql.DB) ([]eventTotal, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	preparedStmt := formulateEventTotalDataPreparedStatement()

	stmt, err := db.Prepare(preparedStmt)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(c.lastEventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id float64
	for rows.Next() {
		var et eventTotal

		err := rows.Scan(&et.count, &et.dag, &et.task, &et.event, &id)
		if err != nil {
			return nil, err
		}

		c.lastEventID = math.Max(id, c.lastEventID)

		if c.data[et.dag] == nil {
			c.data[et.dag] = make(map[string]map[string]float64)
		}
		if c.data[et.dag][et.task] == nil {
			c.data[et.dag][et.task] = make(map[string]float64)
		}
		c.data[et.dag][et.task][et.event] += float64(et.count)
	}

	var etList []eventTotal
	for dag, dagMap := range c.data {
		for task, taskMap := range dagMap {
			for event, total := range taskMap {
				etList = append(etList, eventTotal{
					count: total,
					dag:   dag,
					task:  task,
					event: event,
				})
			}
		}
	}

	return etList, nil
}

func defaultDagRunStates() map[string]float64 {
	return map[string]float64{
		"success": 0.0,
		"failed":  0.0,
		"running": 0.0,
	}
}

func getDagRunStateData(db *sql.DB) ([]dagRunState, error) {

	rows, err := db.Query("SELECT COUNT(*), COALESCE(dag_id, ''), COALESCE(state, '') FROM dag_run GROUP BY dag_id, state")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drsRunStates = make(map[string]map[string]float64)
	for rows.Next() {
		var drs dagRunState

		err := rows.Scan(&drs.count, &drs.dag, &drs.state)
		if err != nil {
			return nil, err
		}

		if drsRunStates[drs.dag] == nil {
			drsRunStates[drs.dag] = defaultDagRunStates()
		}
		drsRunStates[drs.dag][drs.state] = float64(drs.count)
	}

	var drsList []dagRunState
	for dag, dagStates := range drsRunStates {
		for state, count := range dagStates {
			drsList = append(drsList, dagRunState{
				count: count,
				dag:   dag,
				state: state,
			})
		}
	}

	return drsList, nil
}

func getPoolSlotData(db *sql.DB) ([]poolSlot, error) {

	rows, err := db.Query("SELECT poolSlot, slots FROM slot_pool")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var poolSlots []poolSlot
	for rows.Next() {
		var ps poolSlot

		err := rows.Scan(&ps.name, &ps.size)
		if err != nil {
			return nil, err
		}

		poolSlots = append(poolSlots, ps)
	}

	return poolSlots, nil
}
