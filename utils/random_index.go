package utils

import (
	"math/rand"
)

type RandomIndex struct {
	Count         int
	MaxPos        int
	CurrentCount  int
	Pos           int
	Indexes       []int
	IndexLocation map[int]int
	HaltingMode   bool
	Inactive      []int
}

func NewRandomIndex(count int) *RandomIndex {
	var indexes = make([]int, count)
	var indexLocations = map[int]int{}
	for i := 0; i < count; i++ {
		indexes[i] = i
		indexLocations[i] = i
	}
	return &RandomIndex{
		Count:         count,
		MaxPos:        count - 1,
		CurrentCount:  count,
		Pos:           count - 1,
		Indexes:       indexes,
		IndexLocation: indexLocations,
		Inactive:      []int{},
	}
}

func (rpl *RandomIndex) Next(rng *rand.Rand) int {
	if rpl.CurrentCount == 0 {
		if rpl.HaltingMode {
			return -1
		}
		rpl.CurrentCount = rpl.Count
		rpl.Pos = rpl.MaxPos
	}

	var next = RandomIntN(rpl.CurrentCount)
	var value = rpl.Indexes[next]
	if next != rpl.Pos {
		rpl.swap(rpl.Pos, next)
	}
	rpl.Pos -= 1
	rpl.CurrentCount -= 1
	return value
}

func (rpl *RandomIndex) Reset() {
	rpl.CurrentCount = rpl.Count
	rpl.Pos = rpl.MaxPos
}

func (rpl *RandomIndex) RemovePositions(positions ...int) {
	var removedPositions = map[int]bool{}
	for _, value := range positions {
		removedPositions[value] = true
	}
	var newIndexes = make([]int, 0, rpl.Count)
	for _, value := range rpl.Indexes {
		if _, found := removedPositions[value]; !found {
			newIndexes = append(newIndexes, value)
		}
	}
	rpl.Indexes = newIndexes
	rpl.Count = len(newIndexes)
	rpl.MaxPos = rpl.Count - 1
	rpl.Reset()
}

func (rpl *RandomIndex) Shuffle(rng *rand.Rand) {
	rng.Shuffle(len(rpl.Indexes), rpl.swap)
	rpl.Reset()
}

func (rpl *RandomIndex) swap(i, j int) {
	var iVal, jVal = rpl.Indexes[j], rpl.Indexes[i]
	rpl.Indexes[i], rpl.Indexes[j] = iVal, jVal
	rpl.IndexLocation[iVal] = i
	rpl.IndexLocation[jVal] = j
}

func (rpl *RandomIndex) Deactivate(value int) {
	var index = rpl.IndexLocation[value]
	if index < rpl.CurrentCount {
		rpl.swap(rpl.Pos, index)
	}
}
