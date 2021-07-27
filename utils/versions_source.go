package utils

import (
	"gonum.org/v1/gonum/stat/combin"
	"math/bits"
	"math/rand"
	"sync/atomic"
)

var versionId uint64

type VersionSource struct {
	Id             uint64
	Version        Version
	Mask           Version
	minVersionBits int
	maxVersionBits int
	bitCount       int
	RolledVersions []Version
	versionCount   int
	pos            int
	rpl            *RandomIndex
}

func NewVersionSource(version Version, mask Version) *VersionSource {
	bitCount := bits.OnesCount32(uint32(mask))
	vs := &VersionSource{
		Id: atomic.AddUint64(&versionId, 1), Version: version, Mask: mask, bitCount: bitCount,
		minVersionBits: 1, maxVersionBits: bitCount}
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
			for j := 0; j < totalCombinations; j++ {
				tmpMask = 0x0
				for k := 0; k < len(combinations[j]); k++ {
					tmpMask = tmpMask | 1<<bitPositions[combinations[j][k]]
				}
				vs.RolledVersions = append(vs.RolledVersions, vs.Version|tmpMask)
			}
		}
	}
	vs.versionCount = len(vs.RolledVersions)
	vs.rpl = NewRandomIndex(vs.versionCount)
}

func (vs *VersionSource) Retrieve(dest []Version) {
	destCount := len(dest)
	if destCount == 0 {
		return
	}
	for i := 0; i < destCount; i++ {
		if vs.pos >= vs.versionCount {
			vs.pos = 0
		}
		dest[i] = vs.RolledVersions[vs.pos]
		vs.pos += 1
	}
}

func (vs *VersionSource) RNGRetrieve(rng *rand.Rand, dest []Version) {
	var nextPos int
	for pos := range dest {
		nextPos = vs.rpl.Next(rng)
		dest[pos] = vs.RolledVersions[nextPos]
	}
}

func (vs *VersionSource) Shuffle(rng *rand.Rand) {
	vs.rpl.Shuffle(rng)
}

func (vs *VersionSource) shuffler(i, j int) {
	vs.RolledVersions[i], vs.RolledVersions[j] = vs.RolledVersions[j], vs.RolledVersions[i]
}

func (vs *VersionSource) ResetPos() {
	vs.pos = 0
}

func (vs *VersionSource) Len() int {
	return len(vs.RolledVersions)
}

func (vs *VersionSource) Reset() {
	vs.rpl.Reset()
}
