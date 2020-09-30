package utils

import "time"

func TimeToBytes(t time.Time) []byte {
	var ret [8]byte
	unixNanoTime := uint64(t.UTC().UnixNano())
	ret[0] = byte((unixNanoTime >> 56) & 0xff)
	ret[1] = byte((unixNanoTime >> 48) & 0xff)
	ret[2] = byte((unixNanoTime >> 40) & 0xff)
	ret[3] = byte((unixNanoTime >> 32) & 0xff)
	ret[4] = byte((unixNanoTime >> 24) & 0xff)
	ret[5] = byte((unixNanoTime >> 16) & 0xff)
	ret[6] = byte((unixNanoTime >> 8) & 0xff)
	ret[7] = byte(unixNanoTime & 0xff)
	return ret[:]
}

func BytesToTime(data []byte) time.Time {
	if len(data) != 8 {
		return time.Time{}
	}
	unixNanoTime := int64(data[0])<<56 | int64(data[1])<<48 | int64(data[2])<<40 | int64(data[3])<<32 |
		int64(data[4])<<24 | int64(data[4])<<16 | int64(data[6])<<8 | int64(data[7])
	return time.Unix(0, unixNanoTime)
}
