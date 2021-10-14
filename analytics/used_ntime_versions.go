package analytics

import (
	"encoding/json"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
)

type UsedNTime struct {
	rng      *rand.Rand
	NTime    utils.NTime
	Versions []utils.Version
	Index    *utils.RandomIndex
	Count    int
	Ended    bool
}

type UsedNTimes struct {
	rng    *rand.Rand
	NTimes []*UsedNTime
	Index  *utils.RandomIndex
	Count  int
}

func NewUsedNTime(nTime utils.NTime, rawUsedNTime map[utils.Version]uint64) *UsedNTime {
	var ret = &UsedNTime{
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		NTime:    nTime,
		Versions: make([]utils.Version, 0),
	}
	for version := range rawUsedNTime {
		ret.Versions = append(ret.Versions, version)
	}
	ret.Count = len(ret.Versions)
	ret.Index = utils.NewRandomIndex(ret.Count)
	ret.Index.HaltingMode = true
	ret.Index.Shuffle(ret.rng)
	return ret
}

func (ut *UsedNTime) FilterVersions(mask utils.Version) {
	var newVersions = make([]utils.Version, 0)
	var maskFound bool
	for _, version := range ut.Versions {
		var result = version & mask
		if result == mask {
			if result != version || !maskFound {
				newVersions = append(newVersions, version)
				maskFound = true
			}
		}
	}
	ut.Versions = newVersions
	ut.Count = len(ut.Versions)
	if ut.Count > 0 {
		ut.Index = utils.NewRandomIndex(ut.Count)
		ut.Index.HaltingMode = true
		ut.Index.Shuffle(ut.rng)
	}
}

func LoadUsedNTimes(raw map[utils.NTime]map[utils.Version]uint64) *UsedNTimes {
	var ret = &UsedNTimes{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
	for nTime, usedNtimes := range raw {
		ret.NTimes = append(ret.NTimes, NewUsedNTime(nTime, usedNtimes))
	}
	ret.Count = len(ret.NTimes)
	ret.Index = utils.NewRandomIndex(ret.Count)
	ret.Index.Shuffle(ret.rng)
	return ret
}

func (unts *UsedNTimes) FilterVersions(mask utils.Version) {
	var newVersions = make([]*UsedNTime, 0)
	for _, unt := range unts.NTimes {
		unt.FilterVersions(mask)
		if len(unt.Versions) > 0 {
			newVersions = append(newVersions, unt)
		}
	}
	unts.NTimes = newVersions
	unts.Count = len(unts.NTimes)
	unts.Index = utils.NewRandomIndex(unts.Count)
	unts.Index.Shuffle(unts.rng)
}

func (ut *UsedNTime) Reset() {
	ut.Ended = false
	ut.Index.Reset()
}

func (unts *UsedNTimes) Reset() {
	for _, ut := range unts.NTimes {
		ut.Reset()
	}
	unts.Index.Shuffle(unts.rng)
}

func LoadRawUsedNtimes() (*UsedNTimes, error) {
	analyticsFolder := utils.GetSubFolder("analytics")
	f, fileError := os.Open(path.Join(analyticsFolder, "ntimeVersion.json"))
	if fileError != nil {
		return nil, fileError
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.WithError(closeErr).Warn("Error closing file")
		}
	}()
	var decoder = json.NewDecoder(f)
	var usedNTimes = map[utils.NTime]map[utils.Version]uint64{}
	if decodeErr := decoder.Decode(&usedNTimes); decodeErr != nil {
		return nil, decodeErr
	}
	return LoadUsedNTimes(usedNTimes), nil
}

func (ut *UsedNTime) Next(
	allVersions []utils.Version,
	versionsRI *utils.RandomIndex,
	rng *rand.Rand,
) (versions utils.Versions) {
	if ut.Ended {
		ut.Reset()
	}
	for i := 0; i < 4; i++ {
		var nextIndex = ut.Index.Next(ut.rng)
		if nextIndex == -1 {
			versions[i] = allVersions[versionsRI.Next(rng)]
			ut.Ended = true
		} else {
			versions[i] = ut.Versions[nextIndex]
		}
	}
	if ut.Index.CurrentCount == 0 {
		ut.Ended = true
	}
	return
}

func (unts *UsedNTimes) Next(
	allVersions []utils.Version,
	versionsRI *utils.RandomIndex,
	rng *rand.Rand,
) (nTime utils.NTime, versions utils.Versions) {
	var index = unts.Index.Next(unts.rng)
	var ut = unts.NTimes[index]
	nTime = ut.NTime
	versions = ut.Next(allVersions, versionsRI, rng)
	return
}
