package protocol

import "testing"

func TestWorkId(t *testing.T) {
	extWorkId := 0x8765
	midstateMask := (1 << 2) - 1
	t.Logf("%02x", midstateMask)
	workId := extWorkId >> 2
	midstateId := extWorkId & midstateMask
	t.Logf("%02x", 0x10000>>4)
	t.Logf("%02x", workId)
	t.Logf("%02x", midstateId)
}
