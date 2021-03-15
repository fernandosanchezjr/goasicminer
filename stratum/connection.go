package stratum

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type Direction int

var logRPC bool

func init() {
	flag.BoolVar(&logRPC, "log-rpctest", false, "log RPC traffic")
}

type Connection struct {
	conn      *net.TCPConn
	reader    *json.Decoder
	writer    *json.Encoder
	id        uint64
	replyChan chan *protocol.Reply
}

func NewConnection(address string, replyChan chan *protocol.Reply) (*Connection, error) {
	var host, port string
	var addrs []string
	var err error
	if host, port, err = net.SplitHostPort(address); err != nil {
		return nil, err
	}
	if addrs, err = net.LookupHost(host); err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		var rawConn net.Conn
		var conn *net.TCPConn
		var ok bool
		dialer := net.Dialer{Timeout: 1 * time.Second}
		if rawConn, err = dialer.Dial("tcp", fmt.Sprintf("[%s]:%s", addr, port)); err != nil {
			return nil, err
		}
		if conn, ok = rawConn.(*net.TCPConn); !ok {
			return nil, errors.New("invalid connection object")
		}
		if err = conn.SetKeepAlive(true); err != nil {
			return nil, err
		}
		if err = conn.SetKeepAlivePeriod(30 * time.Second); err != nil {
			return nil, err
		}
		c := &Connection{conn: conn, reader: json.NewDecoder(conn), writer: json.NewEncoder(conn), id: 0,
			replyChan: replyChan}
		go c.replyLoop()
		return c, nil
	}
	return nil, fmt.Errorf("No route to %s", address)
}

func (c *Connection) recover() {
	if err := recover(); err != nil {
		log.WithField("error", fmt.Sprint(err)).Error("Call failed")
	}
}

func (c *Connection) Close() error {
	defer c.recover()
	return c.conn.Close()
}

func (c *Connection) NextId() uint64 {
	return atomic.AddUint64(&c.id, 1)
}

func (c *Connection) logRPC(prefix string, value interface{}) {
	if data, err := json.Marshal(value); err != nil {
		log.WithFields(log.Fields{
			"error": fmt.Sprint(err),
		}).Warnln("RPC marshalling error")
	} else {
		log.WithFields(log.Fields{
			"data": data,
		}).Infoln(prefix)
	}
}

func (c *Connection) Call(command protocol.IMethod) error {
	command.SetId(c.NextId())
	if logRPC {
		c.logRPC("RPC out", command)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	doneChan := ctx.Done()
	defer cancel()
	errChan := make(chan error)
	go func() {
		defer c.recover()
		errChan <- c.writer.Encode(command)
	}()
	select {
	case <-doneChan:
		if err := c.Close(); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (c *Connection) readReply(replyChan chan *protocol.Reply, errChan chan error) {
	defer c.recover()
	r := &protocol.Reply{}
	if err := c.reader.Decode(&r); err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			return
		} else {
			log.WithFields(log.Fields{
				"error": fmt.Sprint(err),
			}).Warnln("RPC decode error")
			errChan <- err
		}
	} else {
		if logRPC {
			c.logRPC("RPC in", r)
		}
		replyChan <- r
	}
}

func (c *Connection) replyLoop() {
	var err error
	var replyChan = make(chan *protocol.Reply)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
		doneChan := ctx.Done()
		var errChan = make(chan error)
		go c.readReply(replyChan, errChan)
		err = nil
		select {
		case <-doneChan:
			cancel()
			if err = c.Close(); err != nil {
				log.WithFields(log.Fields{
					"error": fmt.Sprint(err),
				}).Warnln("Connection close error")
			}
			return
		case err = <-errChan:
			cancel()
			if err = c.Close(); err != nil {
				log.WithFields(log.Fields{
					"error": fmt.Sprint(err),
				}).Warnln("Connection read error")
			}
			return
		case reply := <-replyChan:
			cancel()
			c.replyChan <- reply
		}
	}
}
