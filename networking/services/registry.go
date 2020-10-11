package services

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gorpc"
	"github.com/ziutek/ftdi"
	"gonum.org/v1/gonum/optimize"
	"time"
)

func init() {
	var hashRate utils.HashRate
	var difficulty utils.Difficulty
	var nonce32 utils.Nonce32
	var nonce64 utils.Nonce64
	var ntime utils.NTime
	var version utils.Version
	var fields log.Fields
	var level log.Level
	var duration time.Duration
	var interfaceSlice []interface{}
	gorpc.RegisterType(hashRate)
	gorpc.RegisterType(difficulty)
	gorpc.RegisterType(nonce32)
	gorpc.RegisterType(nonce64)
	gorpc.RegisterType(ntime)
	gorpc.RegisterType(version)
	gorpc.RegisterType(fields)
	gorpc.RegisterType(level)
	gorpc.RegisterType(duration)
	gorpc.RegisterType(interfaceSlice)
	gorpc.RegisterType(ftdi.Error{})
	gorpc.RegisterType(optimize.Failure)
}

type Service struct {
	Service interface{}
}

type Registry struct {
	Services map[string]interface{}
}

func NewRegistry() *Registry {
	return &Registry{Services: map[string]interface{}{}}
}

func (r *Registry) AddService(name string, service interface{}) {
	r.Services[name] = service
}
