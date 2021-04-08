package logging

import (
	"flag"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
)

const LogPath = "logs"

var logFile os.File
var logToFile bool

func init() {
	flag.BoolVar(&logToFile, "logToFile", true, "enable logging to file")
}

func getLogFile() *os.File {
	logFolder := utils.GetSubFolder(LogPath)
	f, err := os.OpenFile(path.Join(logFolder, "log.out"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		logrus.Fatal("Error opening log file:", err)
		return nil
	} else {
		return f
	}
}

func exitHandler() {
	_ = logFile.Close()
}

func SetupLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.RegisterExitHandler(exitHandler)
	logrus.SetLevel(logrus.DebugLevel)
	if logToFile {
		logrus.SetOutput(io.MultiWriter(getLogFile(), os.Stdout))
	}
}
