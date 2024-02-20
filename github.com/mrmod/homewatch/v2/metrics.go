package v2

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPort = "2112"
)

var (
	videosCaptured = promauto.NewCounter(prometheus.CounterOpts{
		Name: "videos_captured",
		Help: "The total number of videos captured",
	})
)

type CameraMetrics struct {
	VideosCaptured prometheus.Counter
}

func NewCameraMetrics() *CameraMetrics {
	log.Print("DEBUG: Creating new camera metrics")
	return &CameraMetrics{
		VideosCaptured: videosCaptured,
	}
}
func MetricsHandler() {
	log.Printf("DEBUG: Starting metrics server on port %s", metricsPort)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+metricsPort, nil)
}
