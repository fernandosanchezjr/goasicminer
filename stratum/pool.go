package stratum

import (
	"fmt"
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
	pendingCommands map[uint64]protocol.IMethod
	extraNonce1     uint64
	extraNonce2Len  int
	setDifficulty   *protocol.SetDifficulty
	notify          *protocol.Notify
	workChan        PoolWorkChan
}

func NewPool(config config.Pool, workChan PoolWorkChan) *Pool {
	p := &Pool{
		config:          config,
		status:          Disconnected,
		pendingCommands: make(map[uint64]protocol.IMethod),
		workChan:        workChan,
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

func (p *Pool) Stop() {
	close(p.quit)
	p.wg.Wait()
}

func (p *Pool) String() string {
	return fmt.Sprint(p.config.User, "@", p.config.URL)
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

func (p *Pool) retryTimeout() {
	log.Println("Retrying in", RetryTimeout)
	time.Sleep(RetryTimeout)
}

func (p *Pool) disconnect() {
	if err := p.conn.Close(); err != nil {
		log.Println("Pool", p, "disconnect error:", err)
	}
	p.conn = nil
	p.status = Disconnected
}

func (p *Pool) handleQuit() {
	if p.conn == nil {
		return
	}
	p.disconnect()
	p.quit = nil
}

func (p *Pool) handleDisconnected() {
	if conn, err := NewConnection(p.config.URL); err != nil {
		log.Println("Error connecting to pool", p.config.URL, "-", err)
		p.retryTimeout()
	} else {
		p.conn = conn
		p.status = Connected
	}
}

func (p *Pool) handleConnected() {
	log.Println("Connected to", p)
	subscribe := protocol.NewSubscribe()
	if err := p.conn.Call(subscribe); err != nil {
		log.Println("Pool", p, "subscribe error:", err)
		p.retryTimeout()
		p.disconnect()
	} else {
		p.status = Subscribing
		p.pendingCommands[subscribe.Id] = subscribe
	}
}

func (p *Pool) handleSubscribed() {
	log.Println("Subscribed to", p)
	authorize := protocol.NewAuthorize(p.config.User, p.config.Pass)
	if err := p.conn.Call(authorize); err != nil {
		log.Println("Pool", p, "authorization error:", err)
		p.retryTimeout()
		p.disconnect()
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
			log.Println("Pool", p, "disconnected")
			p.retryTimeout()
		} else if err, ok := err.(net.Error); ok && err.Timeout() {
			return
		} else if err != nil {
			log.Println("Pool", p, "receive error:", err)
		} else if reply != nil {
			if reply.IsMethod() {
				p.handleMethodCall(reply)
			} else {
				p.handleMethodResponse(reply)
			}
		}
	}
}

func (p *Pool) handleMethodResponse(reply *protocol.Reply) {
	method, ok := p.pendingCommands[reply.Id]
	if !ok {
		log.Println("Pool", p, "received unknown reply:", reply)
		return
	}
	switch m := method.(type) {
	case *protocol.Subscribe:
		delete(p.pendingCommands, m.Id)
		if sr, err := protocol.NewSubscribeResponse(reply); err != nil {
			log.Println("Pool", p, "subscribe response error:", err)
		} else {
			p.extraNonce1 = sr.ExtraNonce1
			p.extraNonce2Len = sr.ExtraNonce2Len
			p.status = Subscribed
		}
	case *protocol.Authorize:
		delete(p.pendingCommands, m.Id)
		if ar, err := protocol.NewAuthorizeResponse(reply); err != nil {
			log.Println("Pool", p, "authorize response error:", err)
		} else {
			if ar.Result {
				p.status = Authorized
				log.Println("Pool", p, "authorized")
				p.processWork()
			} else {
				log.Println("Pool", p, "autorization failed")
				p.retryTimeout()
				p.disconnect()
			}
		}
	default:
		log.Println("Pool", p, "received unknown response:", reply)
	}
}

func (p *Pool) handleMethodCall(reply *protocol.Reply) {
	switch reply.MethodName {
	case "mining.set_difficulty":
		p.handleMiningSetDifficulty(reply)
	case "mining.notify":
		p.handleMiningNotify(reply)
	default:
		log.Println("Pool", p, "received unknown remote method:", reply)
	}
}

func (p *Pool) handleMiningSetDifficulty(reply *protocol.Reply) {
	if sd, err := protocol.NewSetDifficulty(reply); err != nil {
		log.Println("Pool", p, "SetDifficulty error:", err)
	} else {
		log.Println("Pool", p, "set difficulty to", sd)
		p.setDifficulty = sd
		p.processWork()
	}
}

func (p *Pool) handleMiningNotify(reply *protocol.Reply) {
	if n, err := protocol.NewNotify(reply); err != nil {
		log.Println("Pool", p, "Notify error:", err)
	} else {
		p.notify = n
		p.processWork()
	}
}

func (p *Pool) processWork() {
	if p.status != Authorized {
		return
	}
	if p.setDifficulty == nil {
		return
	}
	if p.notify == nil {
		return
	}
	p.workChan <- NewPoolWork(p)
}
