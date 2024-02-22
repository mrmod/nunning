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
	}, []string{"camera_name"})
	videosUploaded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "videos_uploaded",
		Help: "The total number of videos uploaded by camera",
	}, []string{"camera_name"})
)

type CameraMetrics struct {
	VideosCaptured *prometheus.CounterVec
	VideosUploaded *prometheus.CounterVec
	VideoEvents    chan string
	UploadEvents   chan string
	trimPrefix     string
}

func NewCameraMetrics(trimPrefix string) *CameraMetrics {
	log.Print("DEBUG: Creating new camera metrics")
	return &CameraMetrics{
		VideosCaptured: videosCaptured,
		VideosUploaded: videosUploaded,
		VideoEvents:    make(chan string, 1),
		UploadEvents:   make(chan string, 1),
		trimPrefix:     trimPrefix,
	}
}

// videoPath: Absolute path to the video file
// trimPrefix: The prefix to remove from the videoPath
func getCameraName(videoPath string, trimPrefix string) string {
	return strings.SplitN(
		strings.TrimPrefix(
			videoPath,
			trimPrefix,
		), string(os.PathSeparator), 2)[0]
}
func (m *CameraMetrics) Handle() {
	defer close(m.VideoEvents)
	defer close(m.UploadEvents)

	go func() {
		log.Print("DEBUG: Starting video event handler")
		for videoEvent := range m.VideoEvents {
			log.Printf("DEBUG: Video event: %s", videoEvent)
			cameraName := getCameraName(videoEvent, m.trimPrefix)
			log.Printf("DEBUG: Camera video event: %s", cameraName)
			m.VideosCaptured.With(prometheus.Labels{"camera_name": cameraName}).Inc()
		}
	}()

	go func() {
		log.Print("DEBUG: Starting upload event handler")
		for uploadEvent := range m.UploadEvents {
			log.Printf("DEBUG: Upload event: %s", uploadEvent)
			cameraName := getCameraName(uploadEvent, m.trimPrefix)
			log.Printf("DEBUG: Camera upload event: %s", cameraName)
			m.VideosUploaded.With(prometheus.Labels{"camera_name": cameraName}).Inc()
		}
	}()

	log.Printf("DEBUG: Starting metrics server on port %s", metricsPort)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+metricsPort, nil)
}
