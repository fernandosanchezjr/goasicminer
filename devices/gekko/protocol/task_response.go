package protocol

type TaskResponse struct {
	Nonce uint32
	JobId int
}

func NewTaskResponse() *TaskResponse {
	return &TaskResponse{}
}

func (tr *TaskResponse) UnmarshalBinary(data []byte) error {
	tr.Nonce = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	tr.JobId = int(data[5])
	return nil
}

func (tr *TaskResponse) BusyResponse() bool {
	return tr.Nonce == 0x83ea0372 || tr.Nonce == 0x09f86be1
}
