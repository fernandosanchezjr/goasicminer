package stratum

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"testing"
	"time"
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
	replyChan := make(chan *protocol.Reply)
	var response *protocol.Reply
	conn, err := NewConnection(poolConfig.URL, replyChan)
	if err != nil {
		t.Fatal(err)
	}
	subscribe := protocol.NewSubscribe()
	if err := conn.Call(subscribe); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		select {
		case <-time.After(time.Second):
			break
		case response = <-replyChan:
			break
		}
		if response != nil {
			t.Log(response.IsMethod(), response)
		} else {
			t.Fatal("no reply")
		}
	}
	configure := protocol.NewConfigure()
	if err := conn.Call(configure); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(time.Second):
		break
	case response = <-replyChan:
		break
	}
	if response != nil {
		t.Log(response.IsMethod(), response)
	} else {
		t.Fatal("no reply")
	}
	auth := protocol.NewAuthorize(poolConfig.User, poolConfig.Pass)
	if err := conn.Call(auth); err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(time.Second):
		break
	case response = <-replyChan:
		break
	}
	if response != nil {
		t.Log(response.IsMethod(), response)
	} else {
		t.Fatal("no reply")
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
}
