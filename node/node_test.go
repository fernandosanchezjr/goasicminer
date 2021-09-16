package node

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNewNode(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Node == nil {
		t.Fatal("No node in cfg file")
	}
	NewNode(cfg.Node)
}

func TestNode_Connect(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Node == nil {
		t.Fatal("No node in cfg file")
	}
	node := NewNode(cfg.Node)
	if err := node.Connect(); err != nil {
		t.Fatal(err)
	}
	node.Disconnect()
}

func TestNode_GetVersions(t *testing.T) {
	t.SkipNow()
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Node == nil {
		t.Fatal("No node in cfg file")
	}
	node := NewNode(cfg.Node)
	if err := node.Connect(); err != nil {
		t.Fatal(err)
	}
	info, infoErr := node.GetInfo()
	if infoErr != nil {
		t.Fatal(err)
	}
	knownVersions := map[utils.Version]bool{}
	for i := info.Blocks; i > 0; i-- {
		header, headerErr := node.GetBlockHeader(i)
		if headerErr != nil {
			t.Fatal(headerErr)
		}
		knownVersions[utils.Version(header.Version)] = true
		if i%1000 == 0 {
			log.Println(i)
		}
	}
	log.Println(len(knownVersions), "versions")
	for version, _ := range knownVersions {
		log.Println(version)
	}
	node.Disconnect()
}

func TestNode_SubmitBlock(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Node == nil {
		t.Fatal("No node in cfg file")
	}
	node := NewNode(cfg.Node)
	if err := node.Connect(); err != nil {
		t.Fatal(err)
	}
	work := <-node.GetWorkChan()
	if submitErr := work.Submit(); submitErr != nil {
		t.Log("error", submitErr)
	} else {
		t.Log("Block submitted")
	}
	node.Disconnect()
}
