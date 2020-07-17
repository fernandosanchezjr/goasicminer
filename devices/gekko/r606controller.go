package gekko

import (
	"bytes"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/protocol"
	"log"
	"time"
)

type R606Controller struct {
	base.IController
	lastReset time.Time
	frequency float64
	chipCount int
}

func NewR606Controller(controller base.IController) *R606Controller {
	return &R606Controller{IController: controller}
}

func (rc *R606Controller) Reset() error {
	if err := rc.PerformReset(); err != nil {
		return err
	}
	if count, err := rc.CountChips(); err != nil {
		return err
	} else {
		log.Println(rc, "found", count, "chips")
	}
	return nil
}

func (rc *R606Controller) PerformReset() error {
	device := rc.USBDevice()
	if _, err := device.Control(64, 0, 0, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 4, 8, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 2, 0, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 0, 2, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 0, 1, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 11, 8434, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 11, 8432, 0, nil); err != nil {
		return err
	}
	if _, err := device.Control(64, 11, 8434, 0, nil); err != nil {
		return err
	}
	return nil
}

func (rc *R606Controller) CountChips() (int, error) {
	var buf bytes.Buffer
	cc := protocol.NewCountChips()
	data, _ := cc.MarshalBinary()
	if err := rc.Write(data); err != nil {
		return 0, err
	}
	time.Sleep(1 * time.Millisecond)
	if err := rc.Read(&buf, 0xff); err != nil {
		return 0, err
	} else {
		ccr := protocol.NewCountChipsResponse()
		if err := ccr.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
			return 0, nil
		} else {
			rc.chipCount = len(ccr.Chips)
			return rc.chipCount, nil
		}
	}
}
