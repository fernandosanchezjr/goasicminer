package protocol

import (
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/utils"
	"math"
)

var setFrequencyMessage = []byte{0x58, 0x09, 0x00, 0x0C, 0x00, 0x50, 0x02, 0x41, 0x00}

type SetFrequency struct {
	data      []byte
	Frequency float64
}

func NewSetFrequency(minFrequency, maxFrequency, frequency float64) *SetFrequency {
	sf := &SetFrequency{data: append([]byte{}, setFrequencyMessage...)}
	if frequency < minFrequency {
		frequency = minFrequency
	} else if frequency > maxFrequency {
		frequency = maxFrequency
	}
	frequency = math.Ceil(100.0*frequency/625.0) * 6.25
	sf.Frequency = frequency
	if frequency < 400 {
		sf.data[5] = byte(frequency*8) / 25
		sf.data[7] = 0x41
	} else {
		sf.data[5] = byte(frequency*4) / 25
		sf.data[7] = 0x21
	}
	utils.BMCRC(sf.data)
	return sf
}

func (sf *SetFrequency) MarshalBinary() ([]byte, error) {
	return sf.data, nil
}