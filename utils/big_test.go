package utils

import (
	log "github.com/sirupsen/logrus"
	"math/big"
	"testing"
)

func Test_DifficultyAssumptions(t *testing.T) {
	resultDiff := big.NewInt(0)
	pdiff := big.NewInt(8192)
	resultDiff.Div(DiffOne, pdiff)
	log.Printf("%x", pdiff.Bytes())
	log.Printf("%d", pdiff.Int64())
	log.Printf("%d", pdiff.Uint64())
}
