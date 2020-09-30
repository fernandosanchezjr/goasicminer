package implementation

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/fernandosanchezjr/goasicminer/backend/services/data"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"time"
)

var BucketNotFound = errors.New("bucket not found")

func GetHostBucket(tx *bbolt.Tx, hostName string) (hostBucket *bbolt.Bucket, err error) {
	if tx.Writable() {
		hostBucket, err = tx.CreateBucketIfNotExists([]byte(hostName))
	} else {
		hostBucket = tx.Bucket([]byte(hostName))
		if hostBucket == nil {
			err = BucketNotFound
		}
	}
	return
}

func GetChildBucket(tx *bbolt.Tx, hostBucket *bbolt.Bucket, name []byte) (rawEventsBucket *bbolt.Bucket, err error) {
	if tx.Writable() {
		rawEventsBucket, err = hostBucket.CreateBucketIfNotExists(name)
	} else {
		rawEventsBucket = hostBucket.Bucket(name)
		if rawEventsBucket == nil {
			err = BucketNotFound
		}
	}
	return
}

func WriteEvent(db *bbolt.DB, hostName string, t time.Time, data []byte) error {
	return db.Update(func(tx *bbolt.Tx) error {
		var err error
		var hostBucket, rawEventsBucket *bbolt.Bucket
		if hostBucket, err = GetHostBucket(tx, hostName); err != nil {
			return err
		}
		if rawEventsBucket, err = GetChildBucket(tx, hostBucket, []byte("rawEvents")); err != nil {
			return err
		}
		if err = rawEventsBucket.Put(utils.TimeToBytes(t), data); err != nil {
			return err
		}
		return nil
	})
}

func GetHostChartBuckets(
	tx *bbolt.Tx,
	hostName string,
	t time.Time,
) (chartBucket *bbolt.Bucket, indexProgressBucket *bbolt.Bucket, err error) {
	var hostBucket, eventsBucket *bbolt.Bucket
	if hostBucket, err = GetHostBucket(tx, hostName); err != nil {
		return
	}
	if _, err = GetChildBucket(tx, hostBucket, []byte("rawEvents")); err != nil {
		return
	}
	if eventsBucket, err = GetChildBucket(tx, hostBucket, []byte("events")); err != nil {
		return
	}
	if t.IsZero() {
		chartBucket = eventsBucket
	} else {
		var yearKey [8]byte
		binary.BigEndian.PutUint64(yearKey[:], uint64(t.Year()))
		if chartBucket, err = GetChildBucket(tx, eventsBucket, yearKey[:]); err != nil {
			return
		}
	}
	indexProgressBucket, err = GetChildBucket(tx, hostBucket, []byte("indexProgress"))
	return
}

func CheckHostIn(db *bbolt.DB, hostName string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		now := time.Now()
		_, _, err := GetHostChartBuckets(tx, hostName, now)
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"hostname": hostName,
			"year":     now.Year(),
		}).Println("Checked in")
		return nil
	})
}

func GetKnownHostNames(db *bbolt.DB) (hostNames []string, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			hostNames = append(hostNames, string(name))
			return nil
		})
	})
	return
}

func GetLastIndexTime(db *bbolt.DB, hostName string, indexName string) (t time.Time, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		_, indexBucket, err := GetHostChartBuckets(tx, hostName, time.Time{})
		if err != nil {
			return err
		}
		value := indexBucket.Get([]byte(indexName))
		if value != nil {
			t = utils.BytesToTime(value)
		}
		return nil
	})
	return
}

func GroupEvents(
	db *bbolt.DB,
	hostName string,
	t time.Time,
) (aggregated map[time.Time][][]byte, lastSeen time.Time, err error) {
	var count int
	var minute time.Time
	aggregated = map[time.Time][][]byte{}
	err = db.View(func(tx *bbolt.Tx) error {
		var err error
		var currMinute time.Time
		var hostBucket, rawEventsBucket *bbolt.Bucket
		if hostBucket, err = GetHostBucket(tx, hostName); err != nil {
			return err
		}
		if rawEventsBucket, err = GetChildBucket(tx, hostBucket, []byte("rawEvents")); err != nil {
			return err
		}
		cursor := rawEventsBucket.Cursor()
		var key, value []byte
		if !t.IsZero() {
			if key, value = cursor.Seek(utils.TimeToBytes(t.Truncate(time.Minute))); key == nil {
				key, value = cursor.First()
			}
		} else {
			key, value = cursor.First()
		}
		for ; key != nil; key, value = cursor.Next() {
			lastSeen = utils.BytesToTime(key)
			currMinute = lastSeen.Truncate(time.Minute)
			if m, found := aggregated[currMinute]; !found {
				aggregated[currMinute] = [][]byte{value}
			} else {
				aggregated[currMinute] = append(m, value)
			}
			if minute.IsZero() || !currMinute.Equal(minute) {
				minute = currMinute
			}
			count += 1
		}
		return nil
	})
	if err != nil {
		return
	}
	log.WithFields(log.Fields{
		"events":   count,
		"hostName": hostName,
		"minutes":  len(aggregated),
	}).Println("Grouped events")
	return
}

func ClearChartAggregations(db *bbolt.DB) error {
	now := time.Now()
	return db.Update(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			if err := b.DeleteBucket([]byte("events")); err != nil {
				return err
			}
			if indexProgressBucket, err := GetChildBucket(tx, b, []byte("indexProgress")); err != nil {
				return err
			} else {
				if err := indexProgressBucket.Delete([]byte("charting")); err != nil {
					return err
				}
			}
			if _, _, err := GetHostChartBuckets(tx, string(name), now); err != nil {
				return err
			}
			return nil
		})
	})
}

func WriteAggregatedMinutes(db *bbolt.DB, hostName string, minutes []*data.Minute) error {
	return db.Update(func(tx *bbolt.Tx) error {
		var err error
		var hostBucket, eventsBucket, yearBucket *bbolt.Bucket
		var currentYear int
		var yearKey [8]byte
		if hostBucket, err = GetHostBucket(tx, hostName); err != nil {
			return err
		}
		if eventsBucket, err = GetChildBucket(tx, hostBucket, []byte("events")); err != nil {
			return err
		}
		for _, m := range minutes {
			if currentYear != m.Time.Year() {
				currentYear = m.Time.Year()
				binary.BigEndian.PutUint64(yearKey[:], uint64(currentYear))
				if yearBucket, err = GetChildBucket(tx, eventsBucket, yearKey[:]); err != nil {
					return err
				}
			}
			var buf bytes.Buffer
			encoder := gob.NewEncoder(&buf)
			if err = encoder.Encode(m); err != nil {
				return err
			}
			if err = yearBucket.Put(utils.TimeToBytes(m.Time), buf.Bytes()); err != nil {
				return err
			}
		}
		return nil
	})
}

func WriteLastIndexTime(db *bbolt.DB, hostName string, indexName string, t time.Time) error {
	return db.Update(func(tx *bbolt.Tx) error {
		_, indexBucket, err := GetHostChartBuckets(tx, hostName, time.Time{})
		if err != nil {
			return err
		}
		return indexBucket.Put([]byte(indexName), utils.TimeToBytes(t))
	})
}

func GetAggregatedMinutes(
	db *bbolt.DB,
	hostName string,
	start time.Time,
	end time.Time,
) (minutes []*data.Minute, err error) {
	var years []time.Time
	yearDiff := end.Year() - start.Year()
	years = append(years, start)
	if yearDiff > 0 {
		for i := 1; i < yearDiff; i++ {
			years = append(years, time.Date(start.Year()+1, 0, 0, 0, 0, 0, 0,
				time.Local))
		}
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		var chartBucket *bbolt.Bucket
		var localErr error
		var cursor *bbolt.Cursor
		var key, value []byte
		var lastTime = start
		for _, year := range years {
			if chartBucket, _, localErr = GetHostChartBuckets(tx, hostName, year); localErr != nil {
				return localErr
			}
			cursor = chartBucket.Cursor()
			key, value = cursor.Seek(utils.TimeToBytes(lastTime))
			for key, value = cursor.Seek(utils.TimeToBytes(lastTime)); key != nil; key, value = cursor.Next() {
				if len(key) == 0 {
					break
				}
				lastTime = utils.BytesToTime(key)
				if lastTime.After(end) {
					break
				}
				var m data.Minute
				buf := bytes.NewBuffer(value)
				decoder := gob.NewDecoder(buf)
				if localErr = decoder.Decode(&m); localErr != nil {
					return localErr
				}
				minutes = append(minutes, &m)
			}
		}
		return nil
	})
	return
}

func IsHostNameKnown(db *bbolt.DB, hostName string) (result bool, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket([]byte(hostName)) != nil {
			result = true
		}
		return nil
	})
	return
}
