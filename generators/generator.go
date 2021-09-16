package generators

import (
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type Generator interface {
	UpdateVersion(versionSource *utils.VersionSource)
	UpdateWork(work *node.Work)
	ExtraNonceFound(extraNonce utils.Nonce64)
	Close()
	GeneratorChan() chan *Generated
	ProgressChan() chan utils.Nonce64
}
