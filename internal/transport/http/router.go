package transporthttp

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
)

var requestCounter uint64

func NewRouter(
	logger *slog.Logger,
	healthHandler *httphandlers.HealthHandler,
	taskHandler *httphandlers.TaskHandler,
	templateHandler *httphandlers.TemplateHandler,
	docsHandler *swaggerdocs.Handler,
) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(requestIDMiddleware)
	router.Use(loggingMiddleware(logger))

	router.HandleFunc("/healthz", healthHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete)

	api.HandleFunc("/task-templates", templateHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/task-templates", templateHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/task-templates/{id:[0-9]+}", templateHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/task-templates/{id:[0-9]+}", templateHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/task-templates/{id:[0-9]+}", templateHandler.Delete).Methods(http.MethodDelete)

	return router
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := buildRequestID()
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			logger.Info(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"request_id", recorder.Header().Get("X-Request-ID"),
				"status", recorder.status,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func buildRequestID() string {
	return time.Now().UTC().Format("20060102T150405.000000000Z07:00") + "-" + strconv.FormatUint(atomic.AddUint64(&requestCounter, 1), 10)
}
