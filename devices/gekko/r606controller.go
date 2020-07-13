package gekko

import (
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/ziutek/ftdi"
	"log"
	"math"
	"time"
)

const (
	R606MinFrequency   = 50.0
	R606MaxFrequency   = 1200.0
	R606ReadChunkSize  = 64
	R606WriteChunkSize = 64
)

type R606Controller struct {
	base.IController
	lastReset time.Time
	frequency float64
}

func NewR606Controller(controller base.IController) *R606Controller {
	return &R606Controller{IController: controller}
}

func (rc *R606Controller) Reset() error {
	if err := rc.IController.Reset(); err != nil {
		return err
	}
	connection := rc.Connection()
	if err := connection.Reset(); err != nil {
		return err
	}
	if err := connection.SetBitmode(0, ftdi.ModeReset); err != nil {
		return err
	}
	if err := connection.SetReadChunkSize(R606ReadChunkSize); err != nil {
		return err
	}
	if err := connection.SetWriteChunkSize(R606WriteChunkSize); err != nil {
		return err
	}
	if err := connection.SetBitmode(8, ftdi.ModeSyncBB); err != nil {
		return err
	}
	if err := connection.SetBaudrate(0xff00); err != nil {
		return err
	}
	if err := connection.SetFlowControl(ftdi.FlowCtrlDisable); err != nil {
		return err
	}
	if err := connection.PurgeReadBuffer(); err != nil {
		return err
	}
	if err := connection.PurgeWriteBuffer(); err != nil {
		return err
	}
	if err := connection.SetBitmode(0xf2, ftdi.ModeCBUS); err != nil {
		return err
	}
	time.Sleep(30 * time.Millisecond)
	if err := connection.SetBitmode(0xf0, ftdi.ModeCBUS); err != nil {
		return err
	}
	time.Sleep(30 * time.Millisecond)
	if err := connection.SetBitmode(0xf2, ftdi.ModeCBUS); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	rc.lastReset = time.Now()
	if count, err := rc.CountChips(); err != nil {
		return err
	} else {
		log.Println(rc, "Found", count, "chips")
	}
	if err := rc.SetFrequency(100.0); err != nil {
		return err
	}
	return nil
}

func (rc *R606Controller) CountChips() (int, error) {
	countPayload := []byte{0x54, 0x05, 0x00, 0x00, 0x00}
	payloadLen := len(countPayload)
	var readBuffer [0xff]byte
	connection := rc.Connection()
	if written, err := connection.Write(countPayload); err != nil {
		return 0, err
	} else if written != payloadLen {
		return 0, fmt.Errorf("error writing %d bytes", payloadLen)
	}
	if read, err := connection.Read(readBuffer[:]); err != nil {
		return 0, err
	} else {
		log.Printf("Read %d bytes\n%x", read, readBuffer)
	}
	return 0, nil
}

func (rc *R606Controller) SetFrequency(frequency float64) error {
	if frequency < R606MinFrequency {
		frequency = R606MinFrequency
	}
	if frequency > R606MaxFrequency {
		frequency = R606MaxFrequency
	}
	if frequency == rc.frequency {
		return nil
	}
	frequencyPayload := []byte{0x58, 0x09, 0x00, 0x0C, 0x00, 0x50, 0x02, 0x41, 0x00}
	payloadLen := len(frequencyPayload)
	frequency = math.Ceil(100.0*(frequency)/625.0) * 6.25
	if frequency < 400.0 {
		frequencyPayload[7] = 0x41
		frequencyPayload[5] = byte((frequency * 8) / 25)
	} else {
		frequencyPayload[7] = 0x21
		frequencyPayload[5] = byte((frequency * 4) / 25)
	}
	connection := rc.Connection()
	if written, err := connection.Write(frequencyPayload); err != nil {
		return err
	} else if written != payloadLen {
		return fmt.Errorf("error writing %d bytes", payloadLen)
	}
	return nil
}
