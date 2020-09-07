package protocol

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"time"
)

func Timing(
	chipCount int,
	frequency float64,
	numCores int,
	waitFactor float64,
) (utils.HashRate, time.Duration, time.Duration) {
	hashRate := float64(chipCount) * frequency * float64(numCores) * 1000000.0
	fullScanMicroSeconds := 1000000.0 * (float64(0xffffffff) / hashRate)
	fullscanDuration := time.Duration(fullScanMicroSeconds*1000.0) * time.Nanosecond
	maxTaskWait := time.Duration(waitFactor * float64(fullscanDuration))
	minWait := 1 * time.Microsecond
	maxWait := 3 * fullscanDuration
	if maxTaskWait < minWait {
		maxTaskWait = minWait
	}
	if maxTaskWait > maxWait {
		maxTaskWait = maxWait
	}
	return utils.HashRate(hashRate), fullscanDuration, maxTaskWait
}
