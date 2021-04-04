package utils

import "math/rand"

type RandomIndex struct {
	Count        int
	MaxPos       int
	CurrentCount int
	Pos          int
	Indexes      []int
}

func NewRandomIndex(count int) *RandomIndex {
	var indexes = make([]int, count)
	for i := 0; i < count; i++ {
		indexes[i] = i
	}
	return &RandomIndex{
		Count:        count,
		MaxPos:       count - 1,
		CurrentCount: count,
		Pos:          count - 1,
		Indexes:      indexes,
	}
}

func (rpl *RandomIndex) Next(rng *rand.Rand) int {
	if rpl.CurrentCount == 0 {
		rpl.CurrentCount = rpl.Count
		rpl.Pos = rpl.MaxPos
	}
	var next = rng.Intn(rpl.CurrentCount)
	var value = rpl.Indexes[next]
	if next != rpl.Pos {
		rpl.Indexes[rpl.Pos], rpl.Indexes[next] = rpl.Indexes[next], rpl.Indexes[rpl.Pos]
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
	var newIndexes = make([]int, 0, rpl.Count)
	var match bool
	for _, value := range rpl.Indexes {
		match = false
		for _, position := range positions {
			if value == position {
				match = true
				break
			}
		}
		if !match {
			newIndexes = append(newIndexes, value)
		}
	}
	rpl.Indexes = newIndexes
	rpl.Count = len(newIndexes)
	rpl.MaxPos = rpl.Count - 1
	rpl.Reset()
}
