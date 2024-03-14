package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	v2 "github.com/mrmod/homewatch/v2"
)

var (
	flagConsolidationInterval      = "5m"
	flagIndexEventApiUrl           string
	flagS3VideoBucketUrl           string
	flagS3IndexBucketUrl           string
	flagSyslogServerAddress        = "0.0.0.0:5140"
	flagIndexEventApiAuthorization string
	flagVideoTrimPrefix            = ""
	flagIndexTrimPrefix            = ""

	flagCleanupAllFiles   bool
	flagCleanupIndexFiles bool
	flagCleanupVideoFiles bool
	flagDebug             bool
	flagVerbose           bool

	flagDecodeVideo       bool
	flagEnableEventUpload bool
	flagEnableVideoUpload bool

	flagEnableV2            bool
	flagV2WatchPaths        string
	flagV2EnableMetrics     bool
	flagV2EnableWatchReaper bool
)

func parseFlags() {
	flag.StringVar(&flagSyslogServerAddress, "syslog-server-address", flagSyslogServerAddress, "IP:Port the Syslog server should listen on")
	flag.StringVar(&flagS3VideoBucketUrl, "s3-video-bucket-url", "", "Video Bucket URL like s3://bucket/some/prefix")
	flag.StringVar(&flagS3IndexBucketUrl, "s3-index-bucket-url", "", "Index bucket URL like s3://bucket/some/prefix")
	flag.StringVar(&flagConsolidationInterval, "consolidation-interval", flagConsolidationInterval, "Submits event datatpoints to indexEventApiUrl after each interval")
	flag.StringVar(&flagIndexEventApiUrl, "index-event-api-url", "", "URL to post indexed event metrics to")
	flag.StringVar(&flagIndexEventApiAuthorization, "index-event-api-authorization", "", "Authorization header value to send when posting metrics")

	flag.BoolVar(&flagCleanupIndexFiles, "cleanup-index-files", false, "Cleanup index files after reading or trying to read them")
	flag.BoolVar(&flagCleanupVideoFiles, "cleanup-video-files", false, "Cleanup video files after uploading or trying to upload")
	flag.BoolVar(&flagCleanupAllFiles, "cleanup-all-files", false, "Cleanup all index or video files following their indiviudual cleanup rules")
	flag.BoolVar(&flagDebug, "debug", false, "Enable debugging output")
	flag.BoolVar(&flagVerbose, "vvv", false, "Enable verbose trace-level output")
	flag.BoolVar(&flagVerbose, "verbose", false, "Enable verbose trace-level output")

	flag.BoolVar(&flagEnableVideoUpload, "enable-video-upload", false, "When true, upload videos to the provided S3 Video Bucket")
	flag.BoolVar(&flagDecodeVideo, "decode-video", false, "When true, decode videos from DAV to H264 before uploading")
	flag.BoolVar(&flagEnableEventUpload, "enable-event-upload", false, "When true upload events to the IndexEventApiUrl")
	flag.StringVar(&flagVideoTrimPrefix, "video-trim-prefix", "", "Prefix to trim from uploaded videos")
	flag.StringVar(&flagIndexTrimPrefix, "index-trim-prefix", "", "Prefix to trim from uploaded indexes")

	flag.BoolVar(&flagEnableV2, "v2", false, "Enable v2 API")
	flag.StringVar(&flagV2WatchPaths, "v2-watch-paths", "", "Comma separated list of paths to watch for changes")
	flag.BoolVar(&flagV2EnableMetrics, "v2-enable-metrics", false, "Enable prometheus metrics")
	flag.BoolVar(&flagV2EnableWatchReaper, "v2-enable-watch-reaper", false, "Enable watch reaper")
	flag.Parse()
	if len(strings.Split(flagSyslogServerAddress, ":")) != 2 {
		panic(fmt.Sprintf("Invalid syslogserveraddress: %s", flagSyslogServerAddress))
	}

	if flagCleanupAllFiles {
		flagCleanupIndexFiles = true
		flagCleanupVideoFiles = true
	}
}

func debugFlags() {
	log.Printf("SyslogServerAddress: %s", flagSyslogServerAddress)
	log.Printf("S3VideoBucketUrl: %s", flagS3VideoBucketUrl)
	log.Printf("S3IndexBucketUrl: %s", flagS3IndexBucketUrl)
	log.Printf("IndexEventApiUrl: %s", flagIndexEventApiUrl)
	log.Printf("CleanupIndexFiles: %v", flagCleanupIndexFiles)
	log.Printf("CleanupVideoFiles: %v", flagCleanupVideoFiles)
	log.Printf("CleanupAllFiles: %v", flagCleanupAllFiles)
	log.Printf("DecodeVideo: %v", flagDecodeVideo)
	log.Printf("EnableVideoUpload: %v", flagEnableVideoUpload)
	log.Printf("EnableEventUpload: %v", flagEnableEventUpload)
	log.Printf("VideoTrimPrefix: %s", flagVideoTrimPrefix)
	log.Printf("DebugOutput: %v", flagDebug)
	log.Printf("VerboseOutput: %v", flagVerbose)
}
func tryCreateS3Uploader() *S3Uploader {
	// Setup S3 uploader for video files
	if flagEnableVideoUpload && strings.HasPrefix(flagS3VideoBucketUrl, "s3://") {
		log.Printf("DEBUG: Creating S3 uploader for video files to %s", flagS3VideoBucketUrl)
		uploader := NewS3Uploader(DefaultS3Client(), flagS3VideoBucketUrl)
		uploader.TrimLocalPrefix(flagVideoTrimPrefix)

		return uploader
	}
	return nil
}

func main() {

	parseFlags()
	if flagDebug {
		debugFlags()
	}

	if flagEnableV2 {
		var metrics *v2.CameraMetrics
		log.Printf("Starting v2")
		fileEvents := make(chan string, 1)
		uploader := tryCreateS3Uploader()
		if flagV2EnableMetrics {
			// v2.MetricsPort = "2112"
			metrics = v2.NewCameraMetrics(flagVideoTrimPrefix)
			go metrics.Handle()
		}
		if flagV2EnableWatchReaper {
			go v2.WatchReaper()
		}

		go func() {
			for fileEvent := range fileEvents {
				log.Printf("DEBUG: File event: %s", fileEvent)
				if flagV2EnableMetrics && metrics != nil {
					log.Printf("DEBUG: Sending video event to metrics: %s", fileEvent)
					metrics.VideoEvents <- fileEvent
				}

				if uploader != nil {
					go func(videoFilename string) {
						done := make(chan int, 1)
						log.Printf("DEBUG: Uploading file: %s", videoFilename)
						go uploader.UploadFile(videoFilename, done)
						<-done
						log.Printf("DEBUG: Done uploading file: %s", videoFilename)
						metrics.UploadEvents <- videoFilename
					}(fileEvent)
				}
			}
		}()
		v2.Listen(fileEvents, strings.Split(flagV2WatchPaths, ",")...)

		return
	}
	log.Printf("Starting v1")
	syslogServer := NewSyslogServer(flagSyslogServerAddress)
	messageHandler := NewSyslogMessageHandler()

	indexEventHandler := NewFileEventHandler(flagEnableEventUpload, messageHandler.IndexEvents)
	videoEventHandler := NewVideoEventHandler(flagEnableVideoUpload, flagDecodeVideo, messageHandler.VideoEvents)

	// Setup S3 Uploader for index files
	if flagEnableEventUpload && strings.HasPrefix(flagS3IndexBucketUrl, "s3://") {
		uploader := NewS3Uploader(DefaultS3Client(), flagS3IndexBucketUrl)
		uploader.TrimLocalPrefix(flagIndexTrimPrefix)

		indexEventHandler.AddUploader(uploader)
	}

	// Setup S3 uploader for video files
	if flagEnableVideoUpload && strings.HasPrefix(flagS3VideoBucketUrl, "s3://") {
		uploader := NewS3Uploader(DefaultS3Client(), flagS3VideoBucketUrl)
		uploader.TrimLocalPrefix(flagVideoTrimPrefix)

		videoEventHandler.AddUploader(uploader)

	}
	go videoEventHandler.Listen()
	go indexEventHandler.Listen()
	go messageHandler.Run()
	go syslogServer.Serve(messageHandler.Messages)

	signals := make(chan os.Signal, 1)
	go func() {
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	}()
	defer close(signals)
	<-signals
	indexEventHandler.Stop()
	syslogServer.Stop()
	log.Printf("Shutdown server")
}
