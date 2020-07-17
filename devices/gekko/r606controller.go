package gekko

import (
	"bytes"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/protocol"
	"log"
	"time"
)

const (
	R606BaudDiv = 1
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
	log.Println("Resetting", rc.LongString())
	if err := rc.PerformReset(); err != nil {
		return err
	}
	if count, err := rc.CountChips(); err != nil {
		return err
	} else {
		log.Println(rc.LongString(), "found", count, "chips")
	}
	if err := rc.SendChainInactive(); err != nil {
		return err
	}
	if err := rc.SendChainInactive(); err != nil {
		return err
	}
	if err := rc.SetBaud(); err != nil {
		return err
	}
	log.Println(rc.LongString(), "reset")
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
	time.Sleep(10 * time.Millisecond)
	if err := rc.Read(&buf); err != nil {
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

func (rc *R606Controller) SendChainInactive() error {
	ci := protocol.NewChainInactive()
	cic := protocol.NewChainInactiveChip(rc.chipCount)
	if data, err := ci.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
	for i := 0; i < rc.chipCount; i++ {
		cic.SetCurrentChip(i)
		if data, err := cic.MarshalBinary(); err != nil {
			return err
		} else {
			if err := rc.Write(data); err != nil {
				return err
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	return nil
}

func (rc *R606Controller) SetBaud() error {
	sba := protocol.NewSetBaudA(R606BaudDiv)
	sbb := protocol.NewSetBaudB(R606BaudDiv)
	if data, err := sba.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	if data, err := sbb.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
