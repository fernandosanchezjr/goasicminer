package protocol

import (
	"encoding/hex"
	"fmt"
)

type CountChipsResponse struct {
	Chips []string
}

func NewCountChipsResponse() *CountChipsResponse {
	return &CountChipsResponse{}
}

func (ccr *CountChipsResponse) UnmarshalBinary(data []byte) error {
	if len(data)%7 != 0 {
		return fmt.Errorf("invalid CountChipsResponse length")
	}
	for len(data) > 0 {
		ccr.Chips = append(ccr.Chips, hex.EncodeToString(data[:7]))
		data = data[7:]
	}
	return nil
}
