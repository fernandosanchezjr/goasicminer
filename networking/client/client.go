package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/networking/services"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/gorpc"
	"net"
	"strings"
	"time"
)

const ClientTimeout = time.Second

var ServiceNotFound = errors.New("Service not found")

func isTimeout(err error) bool {
	return strings.Contains(err.Error(), "timeout")
}

type Client struct {
	RpcClient   *gorpc.Client
	Dispatchers map[string]*gorpc.DispatcherClient
}

func NewClient(address string, registry *services.Registry) *Client {
	ipAddrs, err := net.LookupHost(address)
	if err != nil {
		log.WithError(err).Fatal("Could not look up client address")
	}
	if len(ipAddrs) == 0 {
		log.WithField("address", address).Fatal("Could not resolve address to IP")
	}
	tlsConfig := &tls.Config{}
	cl := &Client{
		RpcClient:   gorpc.NewTLSClient(fmt.Sprintf("%s:12000", ipAddrs[0]), tlsConfig),
		Dispatchers: map[string]*gorpc.DispatcherClient{},
	}
	cl.RpcClient.RequestTimeout = ClientTimeout
	cl.RpcClient.LogError = gorpc.NilErrorLogger
	dispatcher := gorpc.NewDispatcher()
	for name, service := range registry.Services {
		dispatcher.AddService(name, service)
		cl.Dispatchers[name] = dispatcher.NewServiceClient(name, cl.RpcClient)
	}
	return cl
}

func (cl *Client) Start() {
	cl.RpcClient.Start()
}

func (cl *Client) Stop() {
	cl.RpcClient.Stop()
}

func (cl *Client) Restart() {
	cl.Stop()
	cl.Start()
}

func (cl *Client) Call(service, funcName string, request interface{}) (interface{}, error) {
	if client, found := cl.Dispatchers[service]; !found {
		return nil, ServiceNotFound
	} else {
		var result interface{}
		var err error
		if result, err = client.Call(funcName, request); err != nil && isTimeout(err) {
			cl.Restart()
		}
		return result, err
	}
}

func (cl *Client) NewBatch(service string) (*gorpc.DispatcherBatch, error) {
	if client, found := cl.Dispatchers[service]; !found {
		return nil, ServiceNotFound
	} else {
		return client.NewBatch(), nil
	}
}

func (cl *Client) CallBatch(batch *gorpc.DispatcherBatch) error {
	var err error
	if err = batch.Call(); err != nil && isTimeout(err) {
		cl.Restart()
	}
	return err
}

func (cl *Client) Send(service, funcName string, request interface{}) error {
	if client, found := cl.Dispatchers[service]; !found {
		return ServiceNotFound
	} else {
		var err error
		if err = client.Send(funcName, request); err != nil && isTimeout(err) {
			cl.Restart()
		}
		return err
	}
}
