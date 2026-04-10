package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Бизнес метрики
	TasksCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_created_total",
			Help: "Total number of tasks created",
		},
	)

	TasksUpdatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_updated_total",
			Help: "Total number of tasks updated",
		},
	)

	TasksDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_deleted_total",
			Help: "Total number of tasks deleted",
		},
	)

	// Метрики базы данных
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	// Метрики ошибок
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"layer", "operation", "error_type"},
	)
)

// RecordHTTPRequest записывает метрики HTTP запроса
func RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	statusCodeStr := strconv.Itoa(statusCode)

	HTTPRequestDuration.WithLabelValues(method, endpoint, statusCodeStr).Observe(duration.Seconds())
	HTTPRequestTotal.WithLabelValues(method, endpoint, statusCodeStr).Inc()
}

// RecordTaskOperation записывает метрики операций с задачами
func RecordTaskOperation(operation string) {
	switch operation {
	case "create":
		TasksCreatedTotal.Inc()
	case "update":
		TasksUpdatedTotal.Inc()
	case "delete":
		TasksDeletedTotal.Inc()
	}
}

// RecordDatabaseQuery записывает метрики запросов к БД
func RecordDatabaseQuery(operation, table string, duration time.Duration) {
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordError записывает метрики ошибок
func RecordError(layer, operation, errorType string) {
	ErrorsTotal.WithLabelValues(layer, operation, errorType).Inc()
}

// UpdateDatabaseConnections обновляет метрику активных соединений с БД
func UpdateDatabaseConnections(count int) {
	DatabaseConnectionsActive.Set(float64(count))
}