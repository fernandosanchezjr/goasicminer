package utils

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"testing"
)

func TestRandomPositionalList(t *testing.T) {
	var next int
	var rng = rand.New(rand.NewSource(RandomInt64()))
	var rpl = NewRandomIndex(10)

	for i := 0; i < 10; i++ {
		next = rpl.Next(rng)
		log.Println(i, next)
	}
}

func TestRandomIndex_Looping(t *testing.T) {
	var next int
	var rng = rand.New(rand.NewSource(RandomInt64()))
	var rpl = NewRandomIndex(10)

	for i := 0; i < 20; i++ {
		next = rpl.Next(rng)
		log.Println(i, next)
	}
}

func TestRandomIndex_RemovePositions(t *testing.T) {
	var next int
	var rng = rand.New(rand.NewSource(RandomInt64()))
	var rpl = NewRandomIndex(10)
	rpl.RemovePositions(1, 5, 9)
	for i := 0; i < 10; i++ {
		next = rpl.Next(rng)
		log.Println(i, next)
	}
}
