package utils

import (
	"gonum.org/v1/gonum/stat/combin"
	"math/bits"
)

type Versions struct {
	version        uint32
	mask           uint32
	versionBits    int
	bitCount       int
	rolledVersions []uint32
	pos            int
}

func NewVersions(version uint32, mask uint32, versionBits int) *Versions {
	bitCount := bits.OnesCount32(mask)
	vs := &Versions{version: version, mask: mask, bitCount: bitCount, versionBits: versionBits}
	vs.init()
	return vs
}

func (vs *Versions) init() {
	var tmpMask uint32
	versionMask := vs.mask
	bitPositions := make([]int, vs.bitCount)
	pos := 0
	for i := 0; i < 32; i++ {
		if (versionMask & 0x1) == 1 {
			bitPositions[pos] = i
			pos += 1
		}
		versionMask = versionMask >> 1
	}
	vs.rolledVersions = []uint32{}
	if vs.bitCount > 0 && vs.versionBits > 0 && vs.bitCount >= vs.versionBits {
		combinations := combin.Combinations(vs.bitCount, vs.versionBits)
		totalCombinations := len(combinations)
		for i := 0; i < totalCombinations; i++ {
			tmpMask = 0x0
			for j := 0; j < len(combinations[i]); j++ {
				tmpMask = tmpMask | 1<<bitPositions[combinations[i][j]]
			}
			vs.rolledVersions = append(vs.rolledVersions, vs.version|tmpMask)
		}
	}
}

func (vs *Versions) Retrieve(dest []uint32) {
	destCount := len(dest)
	rolledCount := len(vs.rolledVersions)
	if destCount == 0 {
		return
	}
	for i := 0; i < destCount; i++ {
		if i == 0 {
			dest[i] = vs.version
		} else {
			dest[i] = vs.rolledVersions[vs.pos]
			vs.pos += 1
			if vs.pos >= rolledCount {
				vs.pos = 0
			}
		}
	}
}
