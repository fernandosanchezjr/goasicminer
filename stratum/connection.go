package stratum

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"net"
	"sync/atomic"
	"time"
)

type Connection struct {
	conn   *net.TCPConn
	reader *json.Decoder
	writer *json.Encoder
	id     uint64
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
			return nil, errors.New("Invalid connection object")
		}
		if err = conn.SetKeepAlive(true); err != nil {
			return nil, err
		}
		if err = conn.SetKeepAlivePeriod(30 * time.Second); err != nil {
			return nil, err
		}
		if err = conn.SetReadBuffer(4096); err != nil {
			return nil, err
		}
		return &Connection{conn: conn, reader: json.NewDecoder(conn), writer: json.NewEncoder(conn), id: 0}, nil
	}
	return nil, fmt.Errorf("No route to %s", address)
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) NextId() uint64 {
	return atomic.AddUint64(&c.id, 1)
}

func (c *Connection) Call(command protocol.IMethod) error {
	command.SetId(c.NextId())
	return c.writer.Encode(command)
}

func (c *Connection) GetReply() (*protocol.Reply, error) {
	r := &protocol.Reply{}
	if err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return nil, err
	}
	if err := c.reader.Decode(&r); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
