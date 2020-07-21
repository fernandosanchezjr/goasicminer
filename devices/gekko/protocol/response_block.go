package protocol

const (
	MaxTaskResponses = 64
)

type ResponseBlock struct {
	Responses []*TaskResponse
	Count     int
}

func NewResponseBlock() *ResponseBlock {
	rb := &ResponseBlock{Responses: make([]*TaskResponse, MaxTaskResponses)}
	for i := 0; i < MaxTaskResponses; i++ {
		rb.Responses[i] = NewTaskResponse()
	}
	return rb
}

func (rb *ResponseBlock) UnmarshalBinary(data []byte) error {
	pos := 0
	for len(data) > 0 {
		if len(data) >= 7 {
			if err := rb.Responses[pos].UnmarshalBinary(data[:7]); err != nil {
				rb.Count = pos
				return err
			}
			data = data[7:]
			pos++
		} else {
			break
		}
	}
	rb.Count = pos
	return nil
}
