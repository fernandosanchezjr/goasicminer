package server

import (
	"crypto/tls"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/networking/certs"
	"github.com/fernandosanchezjr/goasicminer/networking/services"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gorpc"
)

func NewServer(address string, registry *services.Registry) *gorpc.Server {
	dispatcher := gorpc.NewDispatcher()
	for name, service := range registry.Services {
		dispatcher.AddService(name, service)
	}
	cert, err := certs.GetCert("rpc")
	if err != nil {
		log.WithError(err).Fatal("Could not load RPC cert")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server := gorpc.NewTLSServer(fmt.Sprintf("%s:12000", address), dispatcher.NewHandlerFunc(), tlsConfig)
	server.LogError = gorpc.NilErrorLogger
	//server.LogError = func(format string, args ...interface{}){
	//	fmt.Printf(format, args)
	//}
	return server
}
