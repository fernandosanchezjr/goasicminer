package utils

import (
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestDoubleHash(t *testing.T) {
	plainHeaderText := "20000000b6699ea6706da26c88377183fe0a7163a88bb48f0009dbd20000000000000000e700fd569127d4e2b68c794f7a4ee7435e9f88020ae566421009b90eda7715bf5f4db22c171007ea03bf7440"
	plainHeader, err := hex.DecodeString(plainHeaderText)
	if err != nil {
		t.Fatal(err)
	}
	SwapUint32Bytes(plainHeader)
	hash := DoubleHash(plainHeader)
	log.Println("0000000046f016c18a28475a5ac2daa435583156184f7b9089f44fbbb46504e1")
	log.Println(hex.EncodeToString(hash[:]))
}
