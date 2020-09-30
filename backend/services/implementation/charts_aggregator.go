package implementation

import (
	"bytes"
	"encoding/gob"
	"flag"
	"github.com/ReneKroon/ttlcache"
	"github.com/fernandosanchezjr/goasicminer/backend/services/data"
	"github.com/fernandosanchezjr/goasicminer/backend/services/messages"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"os"
	"sync"
	"time"
)

const OperationTimeout = 5 * time.Minute

var clearCharts bool

func init() {
	flag.BoolVar(&clearCharts, "clear-charting", clearCharts, "clear all chart aggregation data")
}

type ChartsAggregator struct {
	db       *bbolt.DB
	quit     chan struct{}
	waiter   sync.WaitGroup
	cache    *ttlcache.Cache
	lastTime time.Time
}

func NewChartsAggregator(db *bbolt.DB) *ChartsAggregator {
	return &ChartsAggregator{
		db:    db,
		cache: ttlcache.NewCache(),
	}
}

func (ca *ChartsAggregator) Start() {
	if ca.quit != nil {
		return
	}
	ca.quit = make(chan struct{})
	ca.waiter.Add(1)
	go ca.mainLoop()
}

func (ca *ChartsAggregator) Stop() {
	if ca.quit == nil {
		return
	}
	close(ca.quit)
	ca.waiter.Wait()
	ca.quit = nil
}

func (ca *ChartsAggregator) mainLoop() {
	if clearCharts {
		ca.ClearAggregations()
	}
	indexTicker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ca.quit:
			indexTicker.Stop()
			ca.waiter.Done()
			return
		case <-indexTicker.C:
			ca.Index()
		}
	}
}

func (ca *ChartsAggregator) ClearAggregations() {
	log.Println("Clearing chart aggregations")
	if err := ClearChartAggregations(ca.db); err != nil {
		log.WithError(err).Error("Could not clear chart aggregations")
	} else {
		log.Println("Cleared chart aggregations")
	}
	os.Exit(0)
}

func (ca *ChartsAggregator) Index() {
	knownHostNames, err := GetKnownHostNames(ca.db)
	if err != nil {
		log.WithError(err).Error("Error getting known hosts")
		return
	}
	log.WithFields(log.Fields{"hostnames": knownHostNames}).Println("Known hosts")
	for _, name := range knownHostNames {
		if _, found := ca.cache.Get(name); !found {
			ca.cache.SetWithTTL(name, true, OperationTimeout)
			go ca.AggregateHost(name)
		}
	}
}

func (ca *ChartsAggregator) GroupEvents(aggregated map[time.Time][][]byte) (minutes []*data.Minute, err error) {
	var entry messages.LogEntry
	var localErr error
	for t, rawData := range aggregated {
		m := data.NewMinute(t)
		for _, raw := range rawData {
			decoder := gob.NewDecoder(bytes.NewBuffer(raw))
			if localErr = decoder.Decode(&entry); localErr != nil {
				log.WithError(err).Error("Error decoding entry")
			} else {
				if entry.Message == "Result" {
					m.Add(entry.Time, entry.Data)
				}
			}
		}
		minutes = append(minutes, m)
	}
	return
}

func (ca *ChartsAggregator) AggregateHost(hostName string) {
	start := time.Now()
	var err error
	var aggregated map[time.Time][][]byte
	var minutes []*data.Minute
	log.WithField("hostName", hostName).Println("Starting chart aggregation")
	if ca.lastTime.IsZero() {
		if ca.lastTime, err = GetLastIndexTime(ca.db, hostName, "charting"); err != nil {
			log.WithFields(log.Fields{"error": err, "hostName": hostName}).Error(
				"Can't retrieve last chart aggregation time")
		}
	}
	log.WithFields(log.Fields{
		"time":     ca.lastTime,
		"hostName": hostName,
	}).Println("Last aggregate chart time")
	if aggregated, ca.lastTime, err = GroupEvents(ca.db, hostName, ca.lastTime); err != nil {
		log.WithFields(log.Fields{"error": err, "hostName": hostName}).Error(
			"Can't gather events")
	} else {
		if minutes, err = ca.GroupEvents(aggregated); err != nil {
			log.WithFields(log.Fields{"error": err, "hostName": hostName}).Error(
				"Can't group events")
		}
		if err == nil {
			if err = WriteAggregatedMinutes(ca.db, hostName, minutes); err != nil {
				log.WithFields(log.Fields{"error": err, "hostName": hostName}).Error(
					"Can't write aggregated minutes")
			}
		}
		if err == nil {
			if err = WriteLastIndexTime(ca.db, hostName, "charting", ca.lastTime); err != nil {
				log.WithFields(log.Fields{"error": err, "hostName": hostName}).Error(
					"Can't write last index time")
			}
		}
	}
	ca.cache.Remove(hostName)
	log.WithFields(log.Fields{
		"hostName": hostName,
		"duration": time.Since(start),
		"lastSeen": ca.lastTime,
		"minutes":  len(minutes),
	}).Println("Finished chart aggregation")
}
