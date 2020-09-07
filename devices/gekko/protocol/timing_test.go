package protocol

import (
	"fmt"
	"testing"
	"time"
)

func TestTiming(t *testing.T) {
	hashRate, fullScan, maxWait := Timing(12, 700, 114, 0.5)
	if fmt.Sprint(hashRate) != "957.6 GH/s" {
		t.Fatal("invalid hash rate")
	}
	if fullScan != time.Duration(4.485137*float64(time.Millisecond)) {
		t.Fatal("invalid scan duration")
	}
	if maxWait != time.Duration(2.242568*float64(time.Millisecond)) {
		t.Fatal("invalid max wait")
	}
}
