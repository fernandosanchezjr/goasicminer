package utils

import (
	"gonum.org/v1/gonum/stat/combin"
	"math/bits"
	"math/rand"
)

type VersionSource struct {
	Version        Version
	Mask           Version
	minVersionBits int
	maxVersionBits int
	bitCount       int
	RolledVersions []Version
	pos            int
}

func NewVersionSource(version Version, mask Version) *VersionSource {
	bitCount := bits.OnesCount32(uint32(mask))
	vs := &VersionSource{Version: version, Mask: mask, bitCount: bitCount, minVersionBits: 1,
		maxVersionBits: bitCount - 1}
	vs.init()
	return vs
}

func (vs *VersionSource) init() {
	tmpMask := vs.Mask
	bitPositions := make([]int, vs.bitCount)
	pos := 0
	for i := 0; i < 32; i++ {
		if (tmpMask & 0x1) == 1 {
			bitPositions[pos] = i
			pos += 1
		}
		tmpMask = tmpMask >> 1
	}
	vs.RolledVersions = []Version{vs.Version}
	if vs.bitCount > 0 && vs.minVersionBits > 0 && vs.bitCount >= vs.maxVersionBits {
		for i := vs.minVersionBits; i < vs.maxVersionBits; i++ {
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

func (vs *VersionSource) Retrieve(dest []Version) {
	destCount := len(dest)
	rolledCount := len(vs.RolledVersions)
	if destCount == 0 {
		return
	}
	for i := 0; i < destCount; i++ {
		if vs.pos >= rolledCount {
			vs.pos = 0
		}
		dest[i] = vs.RolledVersions[vs.pos]
		vs.pos += 1
	}
}

func (vs *VersionSource) RNGRetrieve(rng *rand.Rand, dest []Version) {
	destCount := len(dest)
	rolledCount := len(vs.RolledVersions)
	if destCount == 0 {
		return
	}
	pos := rng.Intn(len(vs.RolledVersions))
	for i := 0; i < destCount; i++ {
		if pos >= rolledCount {
			pos = 0
		}
		dest[i] = vs.RolledVersions[pos]
		pos += 1
	}
}

func (vs *VersionSource) Clone() *VersionSource {
	ret := *vs
	ret.RolledVersions = append([]Version{}, vs.RolledVersions...)
	return &ret
}

func (vs *VersionSource) Shuffle(rng *rand.Rand) {
	rng.Shuffle(len(vs.RolledVersions), func(i, j int) {
		vs.RolledVersions[i], vs.RolledVersions[j] = vs.RolledVersions[j], vs.RolledVersions[i]
	})
}
