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
	var tmpResponse TaskResponse
	var currentResponse *TaskResponse
	rb.Count = 0
	if len(data) == 0 {
		return nil
	}
	for len(data) > 0 {
		if len(data) >= 7 {
			if err := (&tmpResponse).UnmarshalBinary(data[:7]); err != nil {
				return err
			}
			data = data[7:]
			currentResponse = rb.Responses[rb.Count]
			currentResponse.JobId = tmpResponse.JobId
			currentResponse.Nonce = tmpResponse.Nonce
			rb.Count += 1
			if rb.Count >= len(rb.Responses) {
				break
			}
		} else {
			break
		}
	}
	return nil
}
