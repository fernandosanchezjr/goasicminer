package utils

import (
	"gonum.org/v1/gonum/stat/combin"
	"math/bits"
	"math/rand"
	"sync/atomic"
	"time"
)

type Versions struct {
	Version        uint32
	Mask           uint32
	minVersionBits int
	maxVersionBits int
	bitCount       int
	RolledVersions []uint32
	pos            int32
}

func NewVersions(version uint32, mask uint32, minVersionBits int, maxVersionBits int) *Versions {
	bitCount := bits.OnesCount32(mask)
	vs := &Versions{Version: version, Mask: mask, bitCount: bitCount, minVersionBits: minVersionBits,
		maxVersionBits: maxVersionBits}
	vs.init()
	return vs
}

func (vs *Versions) init() {
	var tmpMask uint32
	versionMask := vs.Mask
	bitPositions := make([]int, vs.bitCount)
	pos := 0
	for i := 0; i < 32; i++ {
		if (versionMask & 0x1) == 1 {
			bitPositions[pos] = i
			pos += 1
		}
		versionMask = versionMask >> 1
	}
	vs.RolledVersions = []uint32{vs.Version}
	if vs.maxVersionBits >= vs.bitCount {
		vs.maxVersionBits = vs.bitCount - 1
	}
	if vs.minVersionBits == 1 {
		vs.maxVersionBits += 1
		for i := 0; i < vs.bitCount; i++ {
			vs.RolledVersions = append(vs.RolledVersions, vs.Version|1<<bitPositions[i])
		}
	}
	if vs.bitCount > 0 && vs.minVersionBits > 0 && vs.bitCount > vs.maxVersionBits {
		for i := vs.minVersionBits; i < vs.maxVersionBits+1; i++ {
			combinations := combin.Combinations(vs.bitCount, i)
			totalCombinations := len(combinations)
			for i := 0; i < totalCombinations; i++ {
				tmpMask = 0x0
				for j := 0; j < len(combinations[i]); j++ {
					tmpMask = tmpMask | 1<<bitPositions[combinations[i][j]]
				}
				vs.RolledVersions = append(vs.RolledVersions, vs.Version|tmpMask)
			}
		}
	}
}

func (vs *Versions) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(vs.RolledVersions), func(i, j int) {
		vs.RolledVersions[i], vs.RolledVersions[j] = vs.RolledVersions[j], vs.RolledVersions[i]
	})
}

func (vs *Versions) Retrieve(dest []uint32) {
	destCount := len(dest)
	rolledCount := int32(len(vs.RolledVersions))
	if destCount == 0 {
		return
	}
	for i := 0; i < destCount; i++ {
		dest[i] = vs.RolledVersions[vs.pos]
		if atomic.AddInt32(&vs.pos, 1) >= rolledCount {
			atomic.StoreInt32(&vs.pos, 0)
		}
	}
}
