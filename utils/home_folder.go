package utils

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

var homeFolder string

func init() {
	flag.StringVar(&homeFolder, "home-folder", "~/.goasicminer", "specify home folder")
}

func GetHomeFolder() string {
	if appHomeFolder, err := homedir.Expand(homeFolder); err != nil {
		log.Fatal("Error parsing home folder:", err)
		return ""
	} else {
		if err := os.MkdirAll(appHomeFolder, 0700); err != nil {
			log.Fatal("Could not create", appHomeFolder)
		}
		return appHomeFolder
	}
}

func GetSubFolder(folderPath string) string {
	targetPath := path.Join(GetHomeFolder(), folderPath)
	if err := os.MkdirAll(targetPath, 0700); err != nil {
		log.WithError(err).Fatal("Could not create", targetPath)
	}
	return targetPath
}
