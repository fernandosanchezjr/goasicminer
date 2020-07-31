package protocol

type Subscribe struct {
	*Method
}

func NewSubscribe() *Subscribe {
	return &Subscribe{&Method{
		Id:         0,
		MethodName: "mining.subscribe",
		Params:     []interface{}{"goasicminer/0.0.1"},
	}}
}
