package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version   = "1.1.1"
	startTime = time.Now()

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

type Response struct {
	Hostname  string            `json:"hostname"`
	Version   string            `json:"version"`
	Revision  string            `json:"revision"`
	Color     string            `json:"color"`
	Message   string            `json:"message"`
	Runtime   string            `json:"runtime"`
	Uptime    string            `json:"uptime"`
	Env       map[string]string `json:"env"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.HandleFunc("/", metricsMiddleware(homeHandler)).Methods("GET")
	r.HandleFunc("/healthz", healthHandler).Methods("GET")
	r.HandleFunc("/readyz", healthHandler).Methods("GET")
	r.HandleFunc("/version", metricsMiddleware(versionHandler)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "9898"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	resp := Response{
		Hostname: hostname,
		Version:  version,
		Revision: os.Getenv("REVISION"),
		Color:    getEnvOrDefault("COLOR", "#34577c"),
		Message:  "Testing metrics-based canary deployment v1.1.1 with traffic",
		Runtime:  fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		Uptime:   time.Since(startTime).String(),
		Env: map[string]string{
			"ENVIRONMENT": getEnvOrDefault("ENVIRONMENT", "unknown"),
			"LOG_LEVEL":   getEnvOrDefault("LOG_LEVEL", "info"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": version,
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
