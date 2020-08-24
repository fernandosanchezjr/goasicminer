package base

import (
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/google/gousb"
	"log"
	"time"
)

var logDeviceTraffic bool

func init() {
	flag.BoolVar(&logDeviceTraffic, "log-device-traffic", false, "log device traffic")
}

type IController interface {
	String() string
	LongString() string
	Close()
	Driver() IDriver
	USBDevice() *gousb.Device
	InEndpoint() *gousb.InEndpoint
	OutEndpoint() *gousb.OutEndpoint
	Equals(other IController) bool
	Initialize() error
	Reset() error
	Write(data []byte) error
	Read(buf *bytes.Buffer) error
	ReadTimeout(buf *bytes.Buffer, timeout time.Duration) error
	UpdateWork(work *stratum.Work)
	WorkChannel() stratum.PoolWorkChan
}

type Controller struct {
	*gousb.Device
	driver            IDriver
	done              func()
	iface             *gousb.Interface
	inEndpointNumber  int
	outEndpointNumber int
	inEndpoint        *gousb.InEndpoint
	outEndpoint       *gousb.OutEndpoint
	readBuffer        []byte
	serialNumber      string
	workChan          stratum.PoolWorkChan
}

func NewController(driver IDriver, device *gousb.Device, inEndpoint, outEndpoint int) *Controller {
	return &Controller{Device: device, driver: driver, done: nil, inEndpointNumber: inEndpoint,
		outEndpointNumber: outEndpoint, workChan: make(stratum.PoolWorkChan, 1)}
}

func (c *Controller) String() string {
	return c.serialNumber
}

func (c *Controller) LongString() string {
	return fmt.Sprintf("%s %s", c.driver, c)
}

func (c *Controller) Close() {
	if c.done != nil {
		c.done()
		c.done = nil
	}
	if c.iface != nil {
		c.iface.Close()
	}
	if err := c.Device.Close(); err != nil {
		log.Printf("Error closing %s: %v", c.LongString(), err)
	}
}

func (c *Controller) Driver() IDriver {
	return c.driver
}

func (c *Controller) USBDevice() *gousb.Device {
	return c.Device
}

func (c *Controller) InEndpoint() *gousb.InEndpoint {
	return c.inEndpoint
}

func (c *Controller) OutEndpoint() *gousb.OutEndpoint {
	return c.outEndpoint
}

func (c *Controller) Equals(other IController) bool {
	oud := other.USBDevice()
	return c.Desc.Address == oud.Desc.Address
}

func (c *Controller) Initialize() error {
	var err error
	if err = c.SetAutoDetach(true); err != nil {
		return err
	}
	if c.serialNumber, err = c.SerialNumber(); err != nil {
		return err
	}
	if c.iface, c.done, err = c.DefaultInterface(); err != nil {
		return err
	}
	if c.inEndpoint, err = c.iface.InEndpoint(c.inEndpointNumber); err != nil {
		return err
	} else {
		c.readBuffer = make([]byte, c.inEndpoint.Desc.MaxPacketSize)
	}
	if c.outEndpoint, err = c.iface.OutEndpoint(c.outEndpointNumber); err != nil {
		return err
	}
	return nil
}

func (c *Controller) Reset() error {
	return nil
}

func (c *Controller) Write(data []byte) error {
	outEndpoint := c.OutEndpoint()
	countLen := len(data)
	if logDeviceTraffic {
		log.Println("Writing:", hex.EncodeToString(data))
	}
	if written, err := outEndpoint.Write(data); err != nil {
		return err
	} else if written != countLen {
		return fmt.Errorf("could not write %d bytes", countLen)
	}
	return nil
}

func (c *Controller) Read(buf *bytes.Buffer) error {
	inEndpoint := c.InEndpoint()
	for {
		if read, err := inEndpoint.Read(c.readBuffer); err != nil {
			return err
		} else {
			buf.Write(c.readBuffer[:read])
			if read < c.inEndpoint.Desc.MaxPacketSize {
				if logDeviceTraffic {
					log.Println("Read:", hex.EncodeToString(buf.Bytes()))
				}
				break
			}
		}
	}
	return nil
}

func (c *Controller) ReadTimeout(buf *bytes.Buffer, timeout time.Duration) error {
	inEndpoint := c.InEndpoint()
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	for {
		if read, err := inEndpoint.ReadContext(ctx, c.readBuffer); err != nil {
			if err == gousb.ErrorTimeout || err == gousb.TransferCancelled || err == gousb.ErrorInterrupted {
				return nil
			}
			return err
		} else {
			buf.Write(c.readBuffer[:read])
			if read < c.inEndpoint.Desc.MaxPacketSize {
				if logDeviceTraffic {
					log.Println("Read:", hex.EncodeToString(buf.Bytes()))
				}
				break
			}
		}
	}
	return nil
}

func (c *Controller) UpdateWork(work *stratum.Work) {
	c.workChan <- work
}

func (c *Controller) WorkChannel() stratum.PoolWorkChan {
	return c.workChan
}
