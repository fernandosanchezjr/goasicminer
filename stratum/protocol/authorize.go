package protocol

type Authorize struct {
	*Method
}

func NewAuthorize(user, pass string) *Authorize {
	return &Authorize{&Method{
		Id:         0,
		MethodName: "mining.authorize",
		Params:     []interface{}{user, pass},
	}}
}
