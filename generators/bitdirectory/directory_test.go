package bitdirectory

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestDetail(t *testing.T) {
	var entry, index int
	entry, index = Detail(6)
	log.WithFields(log.Fields{
		"pos":   6,
		"entry": entry,
		"index": index,
	}).Infoln("Detail")
	entry, index = Detail(16)
	log.WithFields(log.Fields{
		"pos":   16,
		"entry": entry,
		"index": index,
	}).Infoln("Detail")
	entry, index = Detail(64)
	log.WithFields(log.Fields{
		"pos":   64,
		"entry": entry,
		"index": index,
	}).Infoln("Detail")
	entry, index = Detail(145)
	log.WithFields(log.Fields{
		"pos":   145,
		"entry": entry,
		"index": index,
	}).Infoln("Detail")
}

func TestOverview(t *testing.T) {
	var pos int
	pos = Overview(0, 6)
	log.WithFields(log.Fields{
		"pos":   pos,
		"entry": 0,
		"index": 6,
	}).Infoln("Detail")
	pos = Overview(1, 8)
	log.WithFields(log.Fields{
		"pos":   pos,
		"entry": 1,
		"index": 8,
	}).Infoln("Detail")
	pos = Overview(2, 24)
	log.WithFields(log.Fields{
		"pos":   pos,
		"entry": 2,
		"index": 24,
	}).Infoln("Detail")
	pos = Overview(5, 9)
	log.WithFields(log.Fields{
		"pos":   pos,
		"entry": 5,
		"index": 9,
	}).Infoln("Detail")
}
