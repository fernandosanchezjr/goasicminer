package stratum

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"testing"
)

func TestProtocol(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Pools) == 0 {
		t.Fatal("No pools in cfg file")
	}
	poolConfig := cfg.Pools[0]
	conn, err := NewConnection(poolConfig.URL)
	if err != nil {
		t.Fatal(err)
	}
	subscribe := protocol.NewSubscribe()
	if err := conn.Call(subscribe); err != nil {
		t.Fatal(err)
	}
	if response, err := conn.GetReply(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(response.IsMethod(), response)
	}
	if response, err := conn.GetReply(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(response.IsMethod(), response)
	}
	if response, err := conn.GetReply(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(response.IsMethod(), response)
	}
	configure := protocol.NewConfigure()
	if err := conn.Call(configure); err != nil {
		t.Fatal(err)
	}
	if response, err := conn.GetReply(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(response.IsMethod(), response)
	}
	auth := protocol.NewAuthorize(poolConfig.User, poolConfig.Pass)
	if err := conn.Call(auth); err != nil {
		t.Fatal(err)
	}
	if response, err := conn.GetReply(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(response.IsMethod(), response)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
}
