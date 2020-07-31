package gekko

import (
	"log"
	"testing"
)

func TestMidstateBytes(t *testing.T) {
	msbs := GetWiresharkMidstates()
	for _, msbs := range msbs {
		log.Printf("%x", msbs)
	}
}
