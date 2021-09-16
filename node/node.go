package node

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	rpcclient "github.com/stevenroose/go-bitcoin-core-rpc"
	"github.com/stevenroose/go-bitcoin-core-rpc/btcjson"
	"math/big"
	"sync"
)

type State int

const (
	Disconnected State = iota
	Connected
)

type Node struct {
	config         *config.Node
	client         *rpcclient.Client
	status         State
	mtx            sync.Mutex
	chainName      string
	walletAddress  btcutil.Address
	pollingExit    chan struct{}
	workChan       chan *Work
	generateChan   chan int
	generateExit   chan struct{}
	log            *log.Entry
	blockChainInfo *btcjson.GetBlockChainInfoResult
	blockTemplate  *btcjson.GetBlockTemplateResult
}

func NewNode(config *config.Node) *Node {
	return &Node{
		config: config, status: Disconnected,
		workChan:     make(chan *Work, 1024),
		generateChan: make(chan int, 1024),
	}
}

func (n *Node) getClient() (*rpcclient.Client, error) {
	if n.config == nil {
		return nil, errors.New("no node configuration found")
	}
	return rpcclient.New(&rpcclient.ConnConfig{
		Host: n.config.URL,
		User: n.config.User,
		Pass: n.config.Pass,
	})
}

func (n *Node) Connect() error {
	n.mtx.Lock()
	if n.status != Disconnected {
		n.mtx.Unlock()
		return nil
	}
	client, err := n.getClient()
	if err != nil {
		n.mtx.Unlock()
		return err
	}
	n.client = client
	n.status = Connected
	n.mtx.Unlock()
	return n.setup()
}

func (n *Node) setup() error {
	n.log = log.WithField("node", n.config.URL)
	info, infoErr := n.GetInfo()
	if infoErr != nil {
		n.status = Disconnected
		n.client = nil
		return infoErr
	}
	n.chainName = fmt.Sprintf("%snet", info.Chain)
	params, paramsErr := n.GetChainParams()
	if paramsErr != nil {
		n.status = Disconnected
		n.client = nil
		return paramsErr
	}
	addr, addrErr := btcutil.DecodeAddress(n.config.Wallet, params)
	if addrErr != nil {
		n.status = Disconnected
		n.client = nil
		return addrErr
	}
	n.log.WithFields(log.Fields{
		"chain":  info.Chain,
		"blocks": info.Blocks,
	}).Println("Node info")
	n.walletAddress = addr
	n.pollingExit = make(chan struct{})
	if !n.config.ClientOnly {
		go n.pollingLoop()
		go n.generateLoop()
	}
	return nil
}

func (n *Node) Disconnect() {
	n.mtx.Lock()
	if n.status == Disconnected {
		n.mtx.Unlock()
		return
	}
	n.mtx.Unlock()
	n.log.Println("Node disconnecting")
	close(n.pollingExit)
	n.client = nil
	n.status = Disconnected
	n.blockTemplate = nil
	n.blockChainInfo = nil
}

func (n *Node) GetInfo() (*btcjson.GetBlockChainInfoResult, error) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	if n.status == Disconnected {
		return nil, nil
	}
	if result, err := n.client.GetBlockChainInfo(); err != nil {
		return nil, err
	} else {
		n.blockChainInfo = result
		return result, err
	}
}

func (n *Node) GetBlockTemplate() (*btcjson.GetBlockTemplateResult, error) {
	n.mtx.Lock()
	if n.status == Disconnected {
		n.mtx.Unlock()
		return nil, nil
	}
	n.mtx.Unlock()
	options := &btcjson.TemplateRequest{
		Capabilities: []string{"longpoll"},
		Rules:        []string{"segwit"},
	}
	if n.blockTemplate != nil {
		options.LongPollID = n.blockTemplate.LongPollID
	}
	if response, err := n.client.GetBlockTemplate(options); err != nil {
		return nil, err
	} else {
		n.blockTemplate = response
		var nBits uint32
		var resultDiff big.Int
		var diff utils.Difficulty
		if data, err := hex.DecodeString(response.Bits); err != nil {
			return nil, err
		} else {
			nBits = binary.BigEndian.Uint32(data)
		}
		utils.CalculateDifficulty(utils.CompactToBig(nBits), &resultDiff)
		diff = utils.Difficulty(resultDiff.Int64())
		n.log.WithFields(log.Fields{
			"height":       response.Height,
			"transactions": len(response.Transactions),
			"previousHash": response.PreviousHash,
			"nTime":        utils.NTime(response.CurTime),
			"minNtime":     utils.NTime(response.MinTime),
			"maxNtime":     utils.NTime(response.MaxTime),
			"mutable":      response.Mutable,
			"difficulty":   diff,
		}).Println("GetBlockTemplate")
		return response, nil
	}
}

func (n *Node) GetBlockHeader(height int32) (*wire.BlockHeader, error) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	if n.status == Disconnected {
		return nil, nil
	}
	hash, hashErr := n.client.GetBlockHash(int64(height))
	if hashErr != nil {
		return nil, hashErr
	}
	return n.client.GetBlockHeader(hash)
}

func (n *Node) GetChainParams() (*chaincfg.Params, error) {
	switch n.chainName {
	case "":
		fallthrough
	case chaincfg.MainNetParams.Name:
		return &chaincfg.MainNetParams, nil
	case chaincfg.TestNet3Params.Name:
		return &chaincfg.TestNet3Params, nil
	case chaincfg.RegressionNetParams.Name:
		return &chaincfg.RegressionNetParams, nil
	case chaincfg.SimNetParams.Name:
		return &chaincfg.SimNetParams, nil
	default:
		return nil, fmt.Errorf("node.GetChainParams: Unknown chain %s", n.chainName)
	}
}

func (n *Node) GetWorkChan() chan *Work {
	return n.workChan
}
