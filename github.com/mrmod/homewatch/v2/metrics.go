package v2

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPort = "2112"
)

var (
	videosCaptured = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "videos_captured",
		Help: "The total number of videos captured",
	},
		[]string{"camera_name"})
)

type CameraMetrics struct {
	VideosCaptured *prometheus.CounterVec
	VideoEvents    chan string
	trimPrefix     string
}

func NewCameraMetrics(trimPrefix string) *CameraMetrics {
	log.Print("DEBUG: Creating new camera metrics")
	return &CameraMetrics{
		VideosCaptured: videosCaptured,
		VideoEvents:    make(chan string, 1),
		trimPrefix:     trimPrefix,
	}
}
func (m *CameraMetrics) Handle() {
	defer close(m.VideoEvents)

	go func() {
		log.Print("DEBUG: Starting video event handler")
		for videoEvent := range m.VideoEvents {
			log.Printf("DEBUG: Video event: %s", videoEvent)
			cameraName := strings.SplitN(
				strings.TrimPrefix(
					videoEvent,
					m.trimPrefix,
				), string(os.PathSeparator), 2)[0]
			log.Printf("DEBUG: Camera video event: %s", cameraName)
			m.VideosCaptured.With(prometheus.Labels{"camera_name": cameraName}).Inc()
		}
	}()

	log.Printf("DEBUG: Starting metrics server on port %s", metricsPort)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+metricsPort, nil)
}
