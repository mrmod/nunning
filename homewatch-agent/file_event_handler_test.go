package main

import (
	"testing"
)

type MockUploader struct {
	T      *testing.T
	MockS3 chan string
}

func (u *MockUploader) UploadFile(filepath string, done chan<- int) {
	u.T.Logf("Uploading %s\n", filepath)
	u.MockS3 <- filepath
	done <- 1
}

func TestFileEventHandler(t *testing.T) {
	mockFilename := "MockFilename"
	fileEvents := make(chan string, 1)
	uploader := &MockUploader{
		T:      t,
		MockS3: make(chan string, 1),
	}
	defer close(fileEvents)
	defer close(uploader.MockS3)
	eventHandler := NewFileEventHandler(true, fileEvents)
	defer eventHandler.Stop()
	eventHandler.AddUploader(uploader)

	go eventHandler.Listen()
	fileEvents <- mockFilename
	uploadedFilename := <-uploader.MockS3

	if uploadedFilename != mockFilename {
		t.Errorf("Expected %s, got %s", mockFilename, uploadedFilename)
	}
}
