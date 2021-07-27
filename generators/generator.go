package generators

import "github.com/fernandosanchezjr/goasicminer/utils"

type Generator interface {
	UpdateVersion(versionSource *utils.VersionSource)
	UpdateWork()
	ExtraNonceFound(extraNonce utils.Nonce64)
	Close()
	GeneratorChan() chan *Generated
	ProgressChan() chan utils.Nonce64
}
