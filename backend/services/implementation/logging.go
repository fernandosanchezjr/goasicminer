package implementation

import (
	"bytes"
	"encoding/gob"
	"flag"
	"github.com/fernandosanchezjr/goasicminer/backend/services/messages"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

var echoRemoteLogs bool

func init() {
	flag.BoolVar(&echoRemoteLogs, "echo-remote-logs", echoRemoteLogs, "echo all miner logs to console")
}

type Logging struct {
	db *bbolt.DB
}

func NewLogging(db *bbolt.DB) *Logging {
	return &Logging{db: db}
}

func (l *Logging) Ingest(rawEntry []byte) {
	var entry messages.LogEntry
	buf := bytes.NewBuffer(rawEntry)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&entry); err != nil {
		log.WithError(err).Error("Error decoding ingested payload")
	}
	if err := WriteEvent(l.db, entry.HostName, entry.Time, rawEntry); err != nil {
		log.WithError(err).Error("Error ingesting log into time series db")
	}
	if echoRemoteLogs {
		log.WithFields(log.Fields{
			"hostName": entry.HostName,
			"level":    entry.Level,
			"time":     entry.Time,
			"data":     entry.Data,
		}).Println(entry.Message)
	}
}
