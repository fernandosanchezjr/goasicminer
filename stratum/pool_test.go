package stratum

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"log"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Pools) == 0 {
		t.Fatal("No pools in cfg file")
	}
	poolConfig := cfg.Pools[0]
	pool := NewPool(poolConfig)
	defer pool.Stop()
	pool.Start()
	select {
	case ps := <-pool.PoolSettings:
		log.Println("Pool settings:", ps)
	case <-time.After(5 * time.Second):
	}
	time.Sleep(10 * time.Second)
}
