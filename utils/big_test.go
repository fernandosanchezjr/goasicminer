package utils

import (
	"log"
	"math/big"
	"testing"
)

func Test_DifficultyAssumptions(t *testing.T) {
	resultDiff := big.NewInt(0)
	pdiff := big.NewInt(8192)
	resultDiff.Div(DiffOne, pdiff)
	log.Printf("%x", resultDiff.Bytes())
	//000000000007fff8000000000000000000000000000000000000000000000000
	//000000000007fff8000000000000000000000000000000000000000000000000
}
