package protocol

import (
	"encoding/hex"
	"testing"
)

func TestCountChipsResponse_UnmarshalBinary(t *testing.T) {
	if data, err := hex.DecodeString("138790000000071387900000000713879000000007138790000000071387900000000713879" +
		"000000007138790000000071387900000000713879000000007138790000000071387900000000713879000000007"); err != nil {
		t.Fatal(err)
	} else {
		ccr := NewCountChipsResponse()
		if err := ccr.UnmarshalBinary(data); err != nil {
			t.Fatal(err)
		}
		if len(ccr.Chips) != 12 {
			t.Fail()
		}
	}
}
