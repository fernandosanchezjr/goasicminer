package generators

import (
	"log"
	"testing"
)

func TestStrategies(t *testing.T) {
	for _, s := range GeneratorStrategies() {
		log.Println(s)
	}
}
