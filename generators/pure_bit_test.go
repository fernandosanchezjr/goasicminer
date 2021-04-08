package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNewPureBit(t *testing.T) {
	vs := utils.NewVersionSource(0x20000000, 0x1fffe000)
	var pb = NewPureBit()
	pb.UpdateVersion(vs)
	pb.Close()
	var generated Generated
	for i := 0; i < 1000; i++ {
		pb.Next(&generated)
		log.WithFields(log.Fields{
			"extraNonce": generated.ExtraNonce2,
			"nTime":      generated.NTime,
			"version0":   generated.Version0,
			"version1":   generated.Version1,
			"version2":   generated.Version2,
			"version3":   generated.Version3,
		}).Infoln("Generated")
	}
}

func BenchmarkNewPureBit(b *testing.B) {
	vs := utils.NewVersionSource(0x20000000, 0x1fffe000)
	var pb = NewPureBit()
	pb.UpdateVersion(vs)
	pb.Close()
	b.ResetTimer()
	var generated Generated
	for i := 0; i < b.N; i++ {
		pb.Next(&generated)
	}
}
