package shim

type Logging struct {
}

func NewLogging() *Logging {
	return &Logging{}
}

func (l *Logging) Ingest(_ []byte) {

}
