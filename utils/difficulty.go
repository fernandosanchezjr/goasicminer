package utils

import (
	"github.com/dustin/go-humanize"
	"strconv"
)

const MaxRawDifficulty = 1000

type Difficulty uint32

func (d Difficulty) String() string {
	if d < MaxRawDifficulty {
		return strconv.FormatUint(uint64(d), 10)
	} else {
		return humanize.SIWithDigits(float64(d), 2, "")
	}
}
