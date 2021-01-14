package version

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type Version struct {
	versions      [4]utils.Version
	versionSource *utils.VersionSource
	rng           *rand.Rand
}

func NewVersion() *Version {
	return &Version{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (v *Version) Next() [4]utils.Version {
	if v.versionSource == nil {
		return v.versions
	}
	v.versionSource.RNGRetrieve(v.rng, v.versions[:])
	return v.versions
}

func (v *Version) Reset(versionSource *utils.VersionSource) {
	if v.versionSource == nil || v.versionSource.Id != versionSource.Id {
		v.versionSource = versionSource
	}
}

func (v *Version) Reseed() {
	v.rng.Seed(utils.RandomInt64())
}

func (v *Version) Shuffle() {
	if v.versionSource != nil {
		v.versionSource.Shuffle(v.rng)
	}
}
