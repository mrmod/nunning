package v2

import (
	"io/fs"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func tryAddPath(w *fsnotify.Watcher, root string) error {
	if err := w.Add(root); err != nil {
		log.Printf("WARN: Error watching '%s': %s", root, err)
		return err
	}
	log.Printf("INFO: Added watch on '%s'", root)
	return nil
}
func AddPaths(w *fsnotify.Watcher, root string) error {
	errorPaths := []string{}
	walkFun := func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			log.Printf("DEBUG: Skipping file %s", path)
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

func Listen(paths ...string) {
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
				log.Printf("event: %#v Op: %s", event, event.Op)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("modified file: %s", event.Name)
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
		// TODO: Track which paths are watched
		// TODO: Should get list of child paths each iteration
		// TODO: Should add child paths if they are not yet watched
		// TODO: Should remove watched paths if they get no events for 24 hours

		if err := AddPaths(watcher, path); err != nil {
			log.Printf("WARN: Error watching %s: %s", path, err)
		}
	}
	<-done
}
