package messages

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type LogEntry struct {
	HostName string
	Data     log.Fields
	Time     time.Time
	Level    log.Level
	Message  string
}
