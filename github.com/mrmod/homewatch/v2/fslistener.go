package v2

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

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
		err = watcher.Add(path)
		if err != nil {
			panic(err)
		}
		log.Printf("Watching: %s", path)
	}
	<-done
}
