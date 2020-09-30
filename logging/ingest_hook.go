package logging

import (
	"bytes"
	"encoding/gob"
	"github.com/fernandosanchezjr/goasicminer/backend/services/messages"
	"github.com/fernandosanchezjr/goasicminer/networking/client"
	log "github.com/sirupsen/logrus"
	"os"
)

type IngestHook struct {
	client   *client.Client
	HostName string
}

func NewIngestHook(client *client.Client) *IngestHook {
	if hostname, err := os.Hostname(); err != nil {
		panic(err)
	} else {
		return &IngestHook{client: client, HostName: hostname}
	}
}

func (ih *IngestHook) Levels() []log.Level {
	return log.AllLevels
}

func (ih *IngestHook) Fire(entry *log.Entry) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(&messages.LogEntry{
		HostName: ih.HostName,
		Data:     entry.Data,
		Time:     entry.Time,
		Level:    entry.Level,
		Message:  entry.Message,
	}); err != nil {
		return err
	}
	return ih.client.Send("Logging", "Ingest", buf.Bytes())
}
