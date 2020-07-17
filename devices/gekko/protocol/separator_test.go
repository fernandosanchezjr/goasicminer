package protocol

import (
	"encoding/hex"
	"testing"
)

func TestMessageSeparator_Search(t *testing.T) {
	if start, end := Separator.Search([]byte{0x01, 0x60, 0x02, 0x03, 0x04, 0x01}); start != 0 && end != 1 {
		t.Fail()
	}
	if start, end := Separator.Search([]byte{0x01}); start != -1 && end != -1 {
		t.Fail()
	}
	if start, end := Separator.Search([]byte{0x01, 0x02}); start != -1 && end != -1 {
		t.Fail()
	}
}

func TestMessageSeparator_Clean(t *testing.T) {
	hexString := "016013879000000007016013870160900000000713879000000160000713879000000007130160879000000007138790016" +
		"000000007138790000000016007138790000000071387016090000000071387900001600000071387900000000701601387900000000" +
		"701600160016001600160016001600160016001600160016001600160016001600160016001600160016001600160016001600160016" +
		"001600160016001600160016001600160016001600160016001600160016001600160016001600160016001600160016001600160016" +
		"00160016001600160016001600160016001600160016001600160016001600160016001600160016001600160"
	if data, err := hex.DecodeString(hexString); err != nil {
		t.Fatal(err)
	} else {
		result := Separator.Clean(data)
		if hex.EncodeToString(result) != "138790000000071387900000000713879000000007138790000000071387900000000713879"+
			"000000007138790000000071387900000000713879000000007138790000000071387900000000713879000000007" {
			t.Fail()
		}
	}
}
