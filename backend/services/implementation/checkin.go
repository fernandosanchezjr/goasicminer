package implementation

import (
	"go.etcd.io/bbolt"
)

type CheckIn struct {
	db *bbolt.DB
}

func NewCheckIn(aggregateDB *bbolt.DB) *CheckIn {
	return &CheckIn{db: aggregateDB}
}

func (c *CheckIn) Host(name string) error {
	return CheckHostIn(c.db, name)
}
