package protocol

var chainInactiveCommand = []byte{0x55, 0x05, 0x00, 0x00, 0x10}

type ChainInactive struct {
}

func NewChainInactive() *ChainInactive {
	return &ChainInactive{}
}

func (c *ChainInactive) MarshalBinary() (data []byte, err error) {
	return chainInactiveCommand, nil
}
