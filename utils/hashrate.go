package utils

import (
	"github.com/dustin/go-humanize"
	"strconv"
)

const MaxRawHashRate = 1000

type HashRate float64

func (h HashRate) String() string {
	if h < MaxRawHashRate {
		return strconv.FormatFloat(float64(h), 'f', -1, 64)
	} else {
		return humanize.SIWithDigits(float64(h), 2, "H/s")
	}
}

func (h HashRate) Fraction(dividend HashRate) float64 {
	return float64(dividend) / float64(h)
}
