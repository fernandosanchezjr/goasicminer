package protocol

import "time"

type IMethod interface {
	SetId(id uint64)
	GetId() uint64
	Age() time.Duration
}

type Method struct {
	Id         uint64        `json:"id"`
	MethodName string        `json:"method"`
	Params     []interface{} `json:"params"`
	Sent       time.Time     `json:"-"`
}

func (m *Method) SetId(id uint64) {
	m.Id = id
	m.Sent = time.Now()
}

func (m *Method) GetId() uint64 {
	return m.Id
}

func (m *Method) Age() time.Duration {
	return time.Since(m.Sent)
}
