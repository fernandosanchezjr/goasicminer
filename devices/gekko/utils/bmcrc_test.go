package utils

import (
	"testing"
)

func TestBMCRC(t *testing.T) {
	request := []byte{0x54, 0x05, 0x00, 0x00, 0x00}
	BMCRC(request)
	if request[len(request)-1] != 0x19 {
		t.Fatal("BMCRC failure!")
	}
}
