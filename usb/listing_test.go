package usb

import (
	"github.com/fernandosanchezjr/goasicminer/usb/drivers"
	"log"
	"testing"

	"github.com/google/gousb"
)

func TestBasicListing(t *testing.T) {
	ctx := gousb.NewContext()
	defer func() {
		_ = ctx.Close()
	}()

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return KnownDevices.Contains(desc.Vendor, desc.Product)
	})

	defer drivers.Cleanup()

	if err != nil {
		log.Fatalf("Error enumerating devices: %v", err)
	}

	for _, dev := range devs {
		// Once the device has been selected from OpenDevices, it is opened
		// and can be interacted with.
		if !drivers.InUse(dev.Desc.Bus, dev.Desc.Address) {
			if driver, err := KnownDevices.FindDriver(dev); err == nil {
				log.Println("Found:", driver.LongName())
				if initErr := driver.Initialize(); initErr != nil {
					log.Fatalf("Error initializing %s: %v", driver.LongName(), initErr)
				}
			} else {
				_ = dev.Close()
				log.Fatalf("Error opening device %d:%d: %v", dev.Desc.Vendor, dev.Desc.Product, err)
			}
		} else {
			_ = dev.Close()
		}
	}
}
