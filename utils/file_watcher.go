package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"path"
	"time"
)

func watcherLoop(filePath string, watcher *fsnotify.Watcher, f func()) {
	var lastEvent = time.Now()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.WithFields(log.Fields{
				"name": event.Name,
				"op":   event.Op,
			}).Info("File watcher")
			if event.Name == filePath &&
				(event.Op == fsnotify.Write || event.Op == fsnotify.Create) &&
				time.Since(lastEvent) >= time.Duration(60*time.Second) {
				lastEvent = time.Now()
				f()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.WithField("error", fmt.Sprint(err)).Error("File watcher")
		}
	}
}

func NewFileWatcher(filePath string, f func()) (*fsnotify.Watcher, error) {
	var watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	go watcherLoop(filePath, watcher, f)
	err = watcher.Add(path.Dir(filePath))
	return watcher, err
}
