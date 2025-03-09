package main

import (
	"testing"
)

func TestGetSourceFromPath(t *testing.T) {
	path := "/mnt/VideoUploads/Camera1/2022-03-06/001/dav/04/04.51.56-04.52.18[M][0@0][0].idx"

	source := GetSourceFromPath(path)

	expectation := "Camera1"
	if source != expectation {
		t.Fatalf("Expected %s, got %s", expectation, source)
	}
}
