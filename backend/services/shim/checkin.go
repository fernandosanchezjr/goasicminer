package shim

type CheckIn struct {
}

func NewCheckIn() *CheckIn {
	return &CheckIn{}
}

func (c *CheckIn) Host(_ string) error {
	return nil
}
