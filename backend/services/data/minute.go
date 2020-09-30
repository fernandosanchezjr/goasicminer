package data

import (
	"encoding/gob"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"time"
)

var MinuteEvents = map[string]interface{}{
	"difficulty":  utils.Difficulty(0),
	"extraNonce2": utils.Nonce64(0),
	"nonce":       utils.Nonce32(0),
	"nTime":       utils.NTime(0),
	"serial":      "",
	"version":     utils.Version(0),
}

type Minute struct {
	Time       time.Time
	EventTimes []time.Time
	Events     map[string][]interface{}
}

func init() {
	var m Minute
	gob.Register(m)
}

func NewMinute(t time.Time) *Minute {
	return &Minute{
		Time:   t,
		Events: map[string][]interface{}{},
	}
}

func (m *Minute) Backfill(count int, value interface{}) (defaultSlice []interface{}) {
	defaultSlice = make([]interface{}, count)
	for i := 0; i < count; i++ {
		defaultSlice[i] = value
	}
	return
}

func (m *Minute) Add(t time.Time, entries map[string]interface{}) {
	populated := map[string]bool{}
	for key, value := range entries {
		if defaultValue, found := MinuteEvents[key]; found {
			populated[key] = true
			if existing, found := m.Events[key]; found {
				m.Events[key] = append(existing, value)
			} else {
				m.Events[key] = append(m.Backfill(len(m.EventTimes), defaultValue), value)
			}
		}
	}
	for key, existing := range m.Events {
		if _, found := populated[key]; !found {
			if defaultValue, found := MinuteEvents[key]; found {
				m.Events[key] = append(existing, defaultValue)
			}
		}
	}
	m.EventTimes = append(m.EventTimes, t)
}
