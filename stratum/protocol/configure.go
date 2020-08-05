package protocol

type Configure struct {
	*Method
}

func NewConfigure() *Configure {
	return &Configure{&Method{
		Id:         0,
		MethodName: "mining.configure",
		Params: []interface{}{
			[]interface{}{"version-rolling"},
			map[string]interface{}{"version-rolling.mask": "ffffffff", "version-rolling.min-bit-count": 4},
		},
	}}
}
