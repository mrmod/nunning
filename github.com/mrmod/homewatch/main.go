package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
}

func main() {
	parseFlags()
	if flagDebug {
		debugFlags()
	}
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