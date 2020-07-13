package base

import (
	"fmt"
	"sync"
)

type IDevice interface {
	String() string
	LongString() string
	Driver() IDriver
	Controller() IController
	Equals(device IDevice) bool
	Close()
	Unregister()
	Initialize() error
	Reset() error
}

type Devices []IDevice

var devicesMtx sync.Mutex
var allDevices Devices

func (ds Devices) Find(device IDevice) bool {
	for _, d := range ds {
		if d.Equals(device) {
			return true
		}
	}
	return false
}

type Device struct {
	driver     IDriver
	controller IController
}

func NewDevice(driver IDriver, controller IController) *Device {
	return &Device{driver: driver, controller: controller}
}

func (d *Device) String() string {
	return d.controller.String()
}

func (d *Device) LongString() string {
	return fmt.Sprintf("%s %s", d.driver, d.controller)
}

func (d *Device) Driver() IDriver {
	return d.driver
}

func (d *Device) Controller() IController {
	return d.controller
}

func (d *Device) Equals(device IDevice) bool {
	return d.controller.Equals(device.Controller())
}

func (d *Device) Close() {
	d.controller.Close()
}

func (d *Device) Unregister() {
	unregisterDevice(d)
}

func (d *Device) Initialize() error {
	err := d.controller.Initialize()
	if err != nil {
		unregisterDevice(d)
		d.Close()
	}
	return err
}

func (d *Device) Reset() error {
	err := d.controller.Reset()
	if err != nil {
		unregisterDevice(d)
		d.Close()
	}
	return err
}

func deviceInUse(device IDevice) bool {
	devicesMtx.Lock()
	defer devicesMtx.Unlock()
	return allDevices.Find(device)
}

func registerDevice(device IDevice) {
	devicesMtx.Lock()
	defer devicesMtx.Unlock()
	if allDevices.Find(device) {
		return
	}
	allDevices = append(allDevices, device)
}

func unregisterDevice(device IDevice) {
	devicesMtx.Lock()
	defer devicesMtx.Unlock()
	if !allDevices.Find(device) {
		return
	}
	var newDevices Devices
	for _, c := range allDevices {
		if !c.Equals(device) {
			newDevices = append(newDevices, c)
		}
	}
	allDevices = newDevices
}

func cleanDevices() {
	devicesMtx.Lock()
	defer devicesMtx.Unlock()
	for _, d := range allDevices {
		d.Close()
	}
	allDevices = []IDevice{}
}

func GetDevices() []IDevice {
	devicesMtx.Lock()
	defer devicesMtx.Unlock()
	return append([]IDevice{}, allDevices...)
}

func GetDriverDevices(driver IDriver) []IDevice {
	devicesMtx.Lock()
	defer devicesMtx.Unlock()
	var foundDevices []IDevice
	for _, device := range allDevices {
		if driver.Equals(device.Driver()) {
			foundDevices = append(foundDevices, device)
		}
	}
	return foundDevices
}
