package main

import "log"

type FileEventHandler struct {
	enableUpload bool
	fileEvents   chan string
	control      chan int
	Uploader     S3FileUploader
}

func NewFileEventHandler(enableUpload bool, fileEvents chan string) *FileEventHandler {
	return &FileEventHandler{
		enableUpload,
		fileEvents,
		make(chan int, 1),
		nil,
	}
}

func (e *FileEventHandler) AddUploader(uploader S3FileUploader) {
	e.Uploader = uploader
}

func (e *FileEventHandler) Listen() {
	defer close(e.control)
	go func() {
		if flagDebug {
			log.Printf("Listening for idx files")
		}
		for filepath := range e.fileEvents {
			if flagDebug {
				log.Printf("File event %s", filepath)
			}

			if e.enableUpload && e.Uploader != nil {
				done := make(chan int, 1)
				go e.Uploader.UploadFile(filepath, done)
				<-done
				if flagDebug {
					log.Printf("Uploaded %s", filepath)
				}
				if flagCleanupAllFiles || flagCleanupIndexFiles {
					tryRemove(filepath)
				}

			}
		}
	}()
	<-e.control
}

func (e *FileEventHandler) Stop() {
	e.control <- 1
}
