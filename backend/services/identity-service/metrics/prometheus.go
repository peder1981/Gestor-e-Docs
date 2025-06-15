package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPRequests conta o número total de requisições HTTP
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_service_http_requests_total",
			Help: "Total de requisições HTTP.",
		},
		[]string{"handler", "method", "status"},
	)

	// HTTPDuration mede a duração das requisições HTTP
	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "identity_service_http_duration_seconds",
			Help:    "Duração das requisições HTTP em segundos.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler", "method"},
	)

	// AuthOperations conta operações de autenticação por tipo e resultado
	AuthOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_service_auth_operations_total",
			Help: "Total de operações de autenticação por tipo.",
		},
		[]string{"operation", "success"},
	)

	// MongoDBOperations conta operações no MongoDB por tipo
	MongoDBOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "identity_service_mongodb_operations_total",
			Help: "Total de operações no MongoDB por tipo.",
		},
		[]string{"operation", "success"},
	)
)

// Init registra as métricas no Prometheus
func Init() {
	// Registrar métricas
	prometheus.MustRegister(HTTPRequests)
	prometheus.MustRegister(HTTPDuration)
	prometheus.MustRegister(AuthOperations)
	prometheus.MustRegister(MongoDBOperations)
}

// PrometheusHandler retorna um handler HTTP para o Prometheus
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// PrometheusMiddleware é um middleware Gin para coletar métricas Prometheus
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		c.Next()

		// Após a execução da requisição
		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		// Registrar métricas
		HTTPRequests.WithLabelValues(path, c.Request.Method, strconv.Itoa(status)).Inc()
		HTTPDuration.WithLabelValues(path, c.Request.Method).Observe(duration)
	}
}
