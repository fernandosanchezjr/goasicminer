package protocol

import "github.com/fernandosanchezjr/goasicminer/utils"

type TaskResponse struct {
	Nonce utils.Nonce32
	JobId int
}

func NewTaskResponse() *TaskResponse {
	return &TaskResponse{}
}

func (tr *TaskResponse) UnmarshalBinary(data []byte) error {
	tr.Nonce = utils.Nonce32(data[0]) | utils.Nonce32(data[1])<<8 | utils.Nonce32(data[2])<<16 | utils.Nonce32(data[3])<<24
	tr.JobId = int(data[5])
	return nil
}

func (tr *TaskResponse) BusyResponse() bool {
	return tr.Nonce == 0x83ea0372 || tr.Nonce == 0x09f86be1
}
