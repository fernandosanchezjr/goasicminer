package protocol

type Method struct {
	Id         uint64        `json:"id"`
	MethodName string        `json:"method"`
	Params     []interface{} `json:"params"`
}

func (m *Method) SetId(id uint64) {
	m.Id = id
}
