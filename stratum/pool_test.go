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
	workChan := make(PoolWorkChan)
	pool := NewPool(poolConfig, workChan)
	defer pool.Stop()
	pool.Start()
	select {
	case ps := <-workChan:
		log.Println("Pool settings:", ps)
		break
	case <-time.After(10 * time.Second):
		break
	}
}
