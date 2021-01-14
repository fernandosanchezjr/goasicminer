package generators

import "github.com/fernandosanchezjr/goasicminer/utils"

type Generated struct {
	ExtraNonce2 utils.Nonce64
	NTime       int
	Version0    utils.Version
	Version1    utils.Version
	Version2    utils.Version
	Version3    utils.Version
}
