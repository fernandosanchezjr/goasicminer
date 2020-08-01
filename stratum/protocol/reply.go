package protocol

import "errors"

type Reply struct {
	Method
	Result interface{}   `json:"result"`
	Error  []interface{} `json:"error"`
}

func (r *Reply) IsMethod() bool {
	return r.Method.MethodName != ""
}

func (r *Reply) HasError() error {
	if len(r.Error) == 0 {
		return nil
	}
	if errorText, ok := r.Error[1].(string); !ok {
		return errors.New("unknown error")
	} else {
		return errors.New(errorText)
	}
}
