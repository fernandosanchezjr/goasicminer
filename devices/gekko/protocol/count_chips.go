package protocol

var countChipsCommand = []byte{0x54, 0x05, 0x00, 0x00, 0x19}

type CountChips struct {
}

func NewCountChips() *CountChips {
	return &CountChips{}
}

func (cc *CountChips) MarshalBinary() (data []byte, err error) {
	return countChipsCommand, nil
}
