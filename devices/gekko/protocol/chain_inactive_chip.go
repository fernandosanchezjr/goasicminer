package protocol

import "github.com/fernandosanchezjr/goasicminer/devices/gekko/utils"

var chainInactiveChipCommand = []byte{0x41, 0x05, 0x00, 0x00, 0x00}

type ChainInactiveChip struct {
	data        []byte
	chipCount   int
	currentChip int
}

func NewChainInactiveChip(chipCount int) *ChainInactiveChip {
	return &ChainInactiveChip{data: append([]byte{}, chainInactiveChipCommand...), chipCount: chipCount}
}

func (cic *ChainInactiveChip) SetCurrentChip(chip int) {
	cic.currentChip = chip
}

func (cic *ChainInactiveChip) MarshalBinary() ([]byte, error) {
	cic.data[2] = byte((0x100 / cic.chipCount) * cic.currentChip)
	utils.BMCRC(cic.data)
	return cic.data, nil
}
