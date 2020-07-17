package protocol

import "github.com/fernandosanchezjr/goasicminer/devices/gekko/utils"

var setBaudBCommand = []byte{0x58, 0x09, 0x00, 0x1C, 0x40, 0x20, 0x99, 0x80, 0x01}

type SetBaudB struct {
	data []byte
}

func NewSetBaudB(baudDiv int) *SetBaudB {
	sbd := &SetBaudB{data: append([]byte{}, setBaudBCommand...)}
	sbd.data[6] = 0x80 | byte(baudDiv)
	utils.BMCRC(sbd.data)
	return sbd
}

func (sbd *SetBaudB) MarshalBinary() ([]byte, error) {
	return sbd.data, nil
}
