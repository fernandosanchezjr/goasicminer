package protocol

type Command interface {
	SetId(id uint64)
}
