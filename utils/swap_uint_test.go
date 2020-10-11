package utils

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestSwapUint64(t *testing.T) {
	result := Nonce64(SwapUint64(0x1234567834127856))
	if result != 0x5678123478563412 {
		log.WithField("result", result).Fatal("Result")
	}
}
