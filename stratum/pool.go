package stratum

import (
	"encoding/hex"
	"github.com/epiclabs-io/elastic"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type PoolState int

const (
	Disconnected PoolState = iota
	Connected
	Subscribing
	Subscribed
	Authorizing
	Authorized
)

const RetryTimeout = 5 * time.Second

type Pool struct {
	config          config.Pool
	quit            chan struct{}
	conn            *Connection
	status          PoolState
	wg              sync.WaitGroup
	pendingCommands map[uint64]interface{}
	extraNonce1     uint64
	extraNonce2Len  int
	difficulty      uint64
	jobId           string
	prevHash        []byte
	coinBase1       []byte
	coinBase2       []byte
	merkleBranch    [][]byte
	version         []byte
	nbits           []byte
	ntime           []byte
	cleanJobs       bool
	PoolSettings    PoolSettingsChan
}

func NewPool(config config.Pool) *Pool {
	p := &Pool{
		config:          config,
		status:          Disconnected,
		pendingCommands: make(map[uint64]interface{}),
		PoolSettings:    make(PoolSettingsChan),
	}
	return p
}

func (p *Pool) Start() {
	if p.quit != nil {
		return
	}
	p.quit = make(chan struct{})
	p.wg.Add(1)
	go p.loop()
}

func (p *Pool) loop() {
	defer p.wg.Done()
	for {
		select {
		case <-p.quit:
			p.handleQuit()
			return
		default:
			switch p.status {
			case Disconnected:
				p.handleDisconnected()
				continue
			case Connected:
				p.handleConnected()
			case Subscribed:
				p.handleSubscribed()
			}
			p.receiveReply()
		}
	}
}

func (p *Pool) handleQuit() {
	if p.conn == nil {
		return
	}
	if err := p.conn.Close(); err != nil {
		log.Println("Error disconnecting from", p.config.URL)
	}
	p.conn = nil
	p.quit = nil
	p.status = Disconnected
}

func (p *Pool) handleDisconnected() {
	if conn, err := NewConnection(p.config.URL); err != nil {
		log.Println("Error connecting to pool", p.config.URL, "-", err)
		log.Println("Retrying in", RetryTimeout)
		time.Sleep(RetryTimeout)
	} else {
		p.conn = conn
		p.status = Connected
	}
}

func (p *Pool) handleConnected() {
	subscribe := protocol.NewSubscribe()
	if err := p.conn.Call(subscribe); err != nil {
		log.Println("Error subscribing to pool", p.config.URL, "-", err)
		log.Println("Retrying in", RetryTimeout)
		time.Sleep(RetryTimeout)
	} else {
		p.status = Subscribing
		p.pendingCommands[subscribe.Id] = subscribe
	}
}

func (p *Pool) handleSubscribed() {
	authorize := protocol.NewAuthorize(p.config.User, p.config.Pass)
	if err := p.conn.Call(authorize); err != nil {
		log.Println("Error authorizing to pool", p.config.URL, "-", err)
		log.Println("Retrying in", RetryTimeout)
		time.Sleep(RetryTimeout)
		if err := p.conn.Close(); err != nil {
			log.Println("Error disconnecting from", p.config.URL)
		}
		p.conn = nil
		p.status = Disconnected
	} else {
		p.status = Authorizing
		p.pendingCommands[authorize.Id] = authorize
	}
}

func (p *Pool) receiveReply() {
	if p.status != Disconnected {
		if reply, err := p.conn.GetReply(); err == io.EOF {
			p.conn = nil
			p.status = Disconnected
			log.Println("Pool disconnected", p.config.URL)
			log.Println("Retrying in", RetryTimeout)
			time.Sleep(RetryTimeout)
		} else if err, ok := err.(net.Error); ok && err.Timeout() {
			return
		} else if err != nil {
			log.Println("Pool", p.config.URL, "error:", err)
		} else if reply != nil {
			if reply.IsMethod() {
				p.handleRemoteMethodCall(reply)
			} else {
				p.handleMethodResponse(reply)
			}
		}
	}
}

func (p *Pool) Stop() {
	close(p.quit)
	p.wg.Wait()
}

func (p *Pool) handleMethodResponse(reply *protocol.Reply) {
	method, ok := p.pendingCommands[reply.Id]
	if !ok {
		log.Println("No command for reply", reply)
		return
	}
	switch m := method.(type) {
	case *protocol.Subscribe:
		delete(p.pendingCommands, m.Id)
		if sr, err := protocol.NewSubscribeResponse(reply); err != nil {
			log.Println("Error decoding Subscribe response:", err)
		} else {
			p.extraNonce1 = sr.ExtraNonce1
			p.extraNonce2Len = sr.ExtraNonce2Len
			p.status = Subscribed
		}
	case *protocol.Authorize:
		delete(p.pendingCommands, m.Id)
		if ar, err := protocol.NewAuthorizeResponse(reply); err != nil {
			log.Println("Error decoding Authorize response:", err)
		} else {
			if ar.Result {
				p.status = Authorized
				log.Println("Successfully authorized with pool", p.config.URL)
			} else {
				log.Println("Error authorizing to pool", p.config.URL, "-", err)
				log.Println("Retrying in", RetryTimeout)
				time.Sleep(RetryTimeout)
				if err := p.conn.Close(); err != nil {
					log.Println("Error disconnecting from", p.config.URL)
				}
				p.conn = nil
				p.status = Disconnected
			}
		}
	default:
		log.Println("Unknown method response:", reply)
	}
}

func (p *Pool) handleRemoteMethodCall(reply *protocol.Reply) {
	switch reply.MethodName {
	case "mining.set_difficulty":
		p.handleMiningSetDifficulty(reply)
	case "mining.notify":
		p.handleMiningNotify(reply)
	default:
		log.Println("Unknown remote method:", reply)
	}
}

func (p *Pool) handleMiningSetDifficulty(reply *protocol.Reply) {
	if len(reply.Params) != 1 {
		log.Println("Invalid params for", reply.MethodName)
		return
	}
	if err := elastic.Set(&p.difficulty, reply.Params[0]); err != nil {
		log.Println("Error decoding mining.set_difficulty response:", err)
	}
}

func (p *Pool) handleMiningNotify(reply *protocol.Reply) {
	if len(reply.Params) != 9 {
		log.Println("Invalid params for", reply.MethodName)
		return
	}
	var prevHash, coinb1, coinb2, version, nbits, ntime string
	var merkleBranch []string
	if err := elastic.Set(&p.jobId, reply.Params[0]); err != nil {
		log.Println("Error decoding mining.notify jobId:", err)
	}
	if err := elastic.Set(&prevHash, reply.Params[1]); err != nil {
		log.Println("Error decoding mining.notify prevHash:", err)
	}
	if data, err := hex.DecodeString(prevHash); err != nil {
		log.Println("Error decoding mining.notify prevHash hex:", err)
	} else {
		p.prevHash = data
	}
	if err := elastic.Set(&coinb1, reply.Params[2]); err != nil {
		log.Println("Error decoding mining.notify coinb1:", err)
	}
	if data, err := hex.DecodeString(coinb1); err != nil {
		log.Println("Error decoding mining.notify coinb1 hex:", err)
	} else {
		p.coinBase1 = data
	}
	if err := elastic.Set(&coinb2, reply.Params[3]); err != nil {
		log.Println("Error decoding mining.notify coinb2:", err)
	}
	if data, err := hex.DecodeString(coinb2); err != nil {
		log.Println("Error decoding mining.notify coinb2 hex:", err)
	} else {
		p.coinBase2 = data
	}
	if err := elastic.Set(&merkleBranch, reply.Params[4]); err != nil {
		log.Println("Error decoding mining.notify merkleBranch:", err)
	}
	p.merkleBranch = make([][]byte, len(merkleBranch))
	for pos, branch := range merkleBranch {
		if data, err := hex.DecodeString(branch); err != nil {
			log.Println("Error decoding mining.notify merkleBranch hex:", err)
		} else {
			p.merkleBranch[pos] = data
		}
	}
	if err := elastic.Set(&version, reply.Params[5]); err != nil {
		log.Println("Error decoding mining.notify version:", err)
	}
	if data, err := hex.DecodeString(version); err != nil {
		log.Println("Error decoding mining.notify version hex:", err)
	} else {
		p.version = data
	}
	if err := elastic.Set(&nbits, reply.Params[6]); err != nil {
		log.Println("Error decoding mining.notify nbits:", err)
	}
	if data, err := hex.DecodeString(nbits); err != nil {
		log.Println("Error decoding mining.notify nbits hex:", err)
	} else {
		p.nbits = data
	}
	if err := elastic.Set(&ntime, reply.Params[7]); err != nil {
		log.Println("Error decoding mining.notify ntime:", err)
	}
	if data, err := hex.DecodeString(ntime); err != nil {
		log.Println("Error decoding mining.notify ntime hex:", err)
	} else {
		p.ntime = data
	}
	if err := elastic.Set(&p.cleanJobs, reply.Params[8]); err != nil {
		log.Println("Error decoding mining.notify cleanJobs:", err)
	}
	p.PoolSettings <- NewPoolSettings(p)
}
