package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"path"
)

func watcherLoop(filePath string, watcher *fsnotify.Watcher, f func()) {
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
				((event.Op&fsnotify.Write) == fsnotify.Write || (event.Op&fsnotify.Create) == fsnotify.Create) {
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
