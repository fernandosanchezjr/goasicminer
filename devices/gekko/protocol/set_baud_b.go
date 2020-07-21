package protocol

import "github.com/fernandosanchezjr/goasicminer/devices/gekko/utils"

var setBaudGateBlockMessage = []byte{0x58, 0x09, 0x00, 0x1C, 0x40, 0x20, 0x99, 0x80, 0x01}

type BaudGateBlock struct {
	data []byte
}

func NewSetBaudGateBlockMessage(baudDiv int) *BaudGateBlock {
	bgb := &BaudGateBlock{data: append([]byte{}, setBaudGateBlockMessage...)}
	bgb.data[6] = 0x80 | byte(baudDiv)
	utils.BMCRC(bgb.data)
	return bgb
}

func (bgb *BaudGateBlock) MarshalBinary() ([]byte, error) {
	return bgb.data, nil
}
