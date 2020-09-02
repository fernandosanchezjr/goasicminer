package stratum

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"log"
	"net"
	"sync/atomic"
	"time"
)

type Direction int

var logRPC bool

func init() {
	flag.BoolVar(&logRPC, "log-rpc", false, "log RPC traffic")
}

type Connection struct {
	conn      *net.TCPConn
	reader    *json.Decoder
	writer    *json.Encoder
	id        uint64
	replyChan chan *protocol.Reply
}

func NewConnection(address string) (*Connection, error) {
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
		if rawConn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%s", addr, port)); err != nil {
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
		//if err = conn.SetReadBuffer(9152); err != nil {
		//	return nil, err
		//}
		c := &Connection{conn: conn, reader: json.NewDecoder(conn), writer: json.NewEncoder(conn), id: 0,
			replyChan: make(chan *protocol.Reply, 256)}
		go c.mainLoop()
		return c, nil
	}
	return nil, fmt.Errorf("No route to %s", address)
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) NextId() uint64 {
	return atomic.AddUint64(&c.id, 1)
}

func (c *Connection) logRPC(prefix string, value interface{}) {
	if data, err := json.Marshal(value); err != nil {
		log.Println("Debug marshaling error:", err)
	} else {
		log.Println(prefix, string(data))
	}
}

func (c *Connection) Call(command protocol.IMethod) error {
	command.SetId(c.NextId())
	if logRPC {
		c.logRPC("RPC out:", command)
	}
	return c.writer.Encode(command)
}

func (c *Connection) mainLoop() {
	for {
		r := &protocol.Reply{}
		if err := c.reader.Decode(&r); err != nil {
			log.Println("Decode err:", err)
		} else {
			if logRPC {
				c.logRPC("RPC in:", r)
			}
			c.replyChan <- r
		}
	}
}

func (c *Connection) GetReply() (*protocol.Reply, error) {
	var reply *protocol.Reply
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	select {
	case <-ctx.Done():
		return nil, nil
	case reply = <-c.replyChan:
		cancel()
		return reply, nil
	}
}
