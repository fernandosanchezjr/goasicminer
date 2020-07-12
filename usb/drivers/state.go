package drivers

import (
	"log"
	"sync"
)

var stateMtx sync.Mutex
var openDrivers []Driver

func RegisterDriver(d Driver) {
	stateMtx.Lock()
	defer stateMtx.Unlock()
	openDrivers = append(openDrivers, d)
}

func UnregisterDriver(d Driver) {
	stateMtx.Lock()
	defer stateMtx.Unlock()
	var newOpenDrivers []Driver
	for _, od := range openDrivers {
		if !od.Equals(d) {
			newOpenDrivers = append(newOpenDrivers, od)
		}
	}
	openDrivers = newOpenDrivers
}

func InUse(bus, address int) bool {
	for _, d := range openDrivers {
		if d.InUse(bus, address) {
			return true
		}
	}
	return false
}

func Cleanup() {
	stateMtx.Lock()
	defer stateMtx.Unlock()
	for _, d := range openDrivers {
		if err := d.Close(); err != nil {
			log.Printf("Error closing %s: %v\n", d.ShortName(), err)
		}
	}
}
