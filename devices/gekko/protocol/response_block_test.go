package protocol

import (
	"encoding/hex"
	"testing"
)

func TestResponseBlock_UnmarshalBinary(t *testing.T) {
	data, err := hex.DecodeString("01607203ea83001787e16bf8090117947203ea83001787e16bf8090117947203ea83001787e16bf" +
		"809011794")
	if err != nil {
		t.Fatal(err)
	}
	rb := NewResponseBlock()
	if err := rb.UnmarshalBinary(Separator.Clean(data)); err != nil {
		t.Fatal(err)
	}
	if rb.Count != 6 {
		t.Fail()
	}
}
