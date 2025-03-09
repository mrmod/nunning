package main

import (
	"io/ioutil"
	"log"
	"path"
	"sync"
)

type VideoEventHandler struct {
	enableUploads bool
	decodeVideos  bool
	videoEvents   chan string
	control       chan int
	Uploader      S3FileUploader
}

func NewVideoEventHandler(enableUploads, decodeVideos bool, videoEvents chan string) *VideoEventHandler {
	return &VideoEventHandler{
		enableUploads,
		decodeVideos,
		videoEvents,
		make(chan int, 1),
		nil,
	}
}

func (v *VideoEventHandler) AddUploader(uploader S3FileUploader) {
	v.Uploader = uploader
}

func (v *VideoEventHandler) Listen() {
	for filepath := range v.videoEvents {
		if v.enableUploads && v.Uploader != nil {
			go func(videofilePath string) {

				wg := &sync.WaitGroup{}
				wg.Add(1)
				go uploadVideo(videofilePath, v.Uploader, wg)
				wg.Wait()

				if flagCleanupVideoFiles || flagCleanupAllFiles {
					_, fn := path.Split(videofilePath)
					log.Printf("Removing %s", fn)
					tryRemove(videofilePath)
				}
			}(filepath)
		}
	}
}
func debugFilepath(filepath string) {
	dirname := path.Dir(filepath)

	log.Printf("Listing %s", dirname)
	infos, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Printf("Unable to list %s: %s", dirname, err)
		return
	}
	for _, info := range infos {
		log.Printf("\t%s", info.Name())
	}
}

func uploadVideo(filepath string, uploader S3FileUploader, wg *sync.WaitGroup) {
	done := make(chan int, 1)
	defer close(done)
	defer wg.Done()

	go uploader.UploadFile(filepath, done)
	for msg := range done {
		switch msg {
		case ErrorUploadingVideoFile:
			log.Printf("Error uploading the video file %s", filepath)
			return
		case ErrorOpeningVideoFile:
			log.Printf("Unable to open video file %s", filepath)
			if flagDebug {
				debugFilepath(filepath)
			}
			return
		case StartUploadVideoFile:
			log.Printf("Started uploading %s", filepath)
		case DoneUploadVideoFile:
			log.Printf("Finished uploading %s", filepath)
			return
		}
	}
}
