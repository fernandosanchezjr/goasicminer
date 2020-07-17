package protocol

import "github.com/fernandosanchezjr/goasicminer/devices/gekko/utils"

var setBaudACommand = []byte{0x58, 0x09, 0x00, 0x1C, 0x00, 0x20, 0x07, 0x00, 0x00}

type SetBaudA struct {
	data []byte
}

func NewSetBaudA(baudDiv int) *SetBaudA {
	sba := &SetBaudA{data: append([]byte{}, setBaudACommand...)}
	sba.data[6] = byte(baudDiv)
	utils.BMCRC(sba.data)
	return sba
}

func (sba *SetBaudA) MarshalBinary() ([]byte, error) {
	return sba.data, nil
}
