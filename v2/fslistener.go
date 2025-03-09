package v2

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatchedPath struct {
	Path           string
	WatchStartTime time.Time
}

type UploadedPath struct {
	Path       string
	UploadTime time.Time
}

var (
	watchedPaths  = map[string]*WatchedPath{}
	uploadedPaths = map[string]*UploadedPath{}
)

// WatchReaper Cleans up watches older than 24 hours
func WatchReaper() {
	log.Printf("DEBUG: Starting watch reaper")
	for {
		time.Sleep(1 * time.Hour)
		for path, watchedPath := range watchedPaths {
			if time.Since(watchedPath.WatchStartTime) > (24 * time.Hour) {
				log.Printf("DEBUG: Removing watch on %s from %s", path, watchedPath.WatchStartTime)
				delete(watchedPaths, path)
			}
		}
	}
}

// UploadReaper Cleans up uploaded files older than 7 days
func UploadReaper() {
	log.Printf("DEBUG: Starting upload reaper")
	for {
		time.Sleep(1 * time.Hour)
		for path, uploadedPath := range uploadedPaths {
			if time.Since(uploadedPath.UploadTime) > (24 * 7 * time.Hour) {
				log.Printf("DEBUG: Removing uploaded file on %s uploaded at %s", path, uploadedPath.UploadTime)
				if err := os.Remove(path); err != nil {
					log.Printf("WARN: Error removing uploaded file %s: %s", path, err)
				}
			}
		}
	}
}

func tryAddPath(w *fsnotify.Watcher, root string) error {
	if err := w.Add(root); err != nil {
		log.Printf("WARN: Error watching '%s': %s", root, err)
		return err
	}
	log.Printf("INFO: Added watch on '%s'", root)
	watchedPaths[root] = &WatchedPath{root, time.Now().UTC()}
	return nil
}

// AddPaths adds child paths of some root path
func AddPaths(w *fsnotify.Watcher, root string) error {
	errorPaths := []string{}
	walkFun := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("WARN: Error walking %s: %s", path, err)
			return err
		}
		if !d.IsDir() {
			// log.Printf("DEBUG: Won't add watcher on file %s", path)
			return nil
		}
		if err := tryAddPath(w, path); err != nil {
			log.Printf("DEBUG: Error walking %s: %s", path, err)
			errorPaths = append(errorPaths, path)
			return err
		}
		log.Printf("DEBUG: Walking %s", path)
		return nil
	}
	return filepath.WalkDir(root, walkFun)
}

// HandleCreate is called when a video file is created by a camera
func HandleCreate(filenames chan string, event fsnotify.Event) {
	log.Printf("DEBUG: Created file: %s", event.Name)
	if strings.HasSuffix(event.Name, ".dav") {
		filenames <- event.Name
	}
}

// Listen starts watching for changes on the given paths
func Listen(createEventFilenames chan string, paths ...string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:

				if !ok {
					return
				}
				// Don't emit events for temporary files
				if !strings.HasSuffix(event.Name, "_") {
					log.Printf("DEBUG: event: %#v Op: %s", event, event.Op)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Add a watcher if this is a path
					if err := AddPaths(watcher, event.Name); err != nil {
						log.Printf("WARN: Error adding path new path %s: %s", event.Name, err)
					}
					//
					HandleCreate(createEventFilenames, event)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("error: %s", err)
			}
		}
	}()
	for _, path := range paths {

		if err := AddPaths(watcher, path); err != nil {
			log.Printf("WARN: Error watching %s: %s", path, err)
		}
	}
	<-done
}
