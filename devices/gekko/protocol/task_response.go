package protocol

type TaskResponse struct {
	Nonce    uint32
	Midstate byte
	JobId    byte
}

func NewTaskResponse() *TaskResponse {
	return &TaskResponse{}
}

func (tr *TaskResponse) UnmarshalBinary(data []byte) error {
	tr.Nonce = uint32(data[3]) | uint32(data[2])<<8 | uint32(data[1])<<16 | uint32(data[0])<<24
	tr.Midstate = data[4]
	tr.JobId = data[5]
	return nil
}

func (tr *TaskResponse) IsBusyWork() bool {
	return tr.Nonce == 0x7203ea83 || tr.Nonce == 0xe16bf809
}
