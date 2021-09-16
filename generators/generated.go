package generators

import (
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type Generated struct {
	Work     *node.Work
	NTime    utils.NTime
	Version0 utils.Version
	Version1 utils.Version
	Version2 utils.Version
	Version3 utils.Version
}
