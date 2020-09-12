package protocol

const (
	MaxTaskResponses = 0x7f
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
	rb.Count = 0
	if len(data) == 0 {
		return nil
	}
	for len(data) > 0 {
		start, end := Separator.Search(data)
		if start != -1 && end != -1 {
			data = data[end+1:]
		}
		if len(data) >= 7 {
			if err := rb.Responses[rb.Count].UnmarshalBinary(data[:7]); err != nil {
				return err
			}
			data = data[7:]
			rb.Count += 1
		} else {
			break
		}
	}
	return nil
}
