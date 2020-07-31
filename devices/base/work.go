package base

import "github.com/fernandosanchezjr/goasicminer/utils"

type Work struct {
	Id       int
	Midstate utils.MidstateBytes
}

type WorkChan chan *Work

func NewWork(id int, midstate utils.MidstateBytes) *Work {
	return &Work{Id: id, Midstate: midstate}
}
