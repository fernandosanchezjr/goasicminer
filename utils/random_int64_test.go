package utils

import "testing"

func TestMaskedRandomInt64(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Log(MaskedRandomInt64())
	}
}
