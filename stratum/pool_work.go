package stratum

import "fmt"

type PoolWork struct {
	ExtraNonce1    uint64
	ExtraNonce2Len int
	Difficulty     uint64
	JobId          string
	PrevHash       []byte
	CoinBase1      []byte
	CoinBase2      []byte
	MerkleBranches [][]byte
	Version        []byte
	Nbits          []byte
	Ntime          []byte
	CleanJobs      bool
	Pool           *Pool
}

type PoolWorkChan chan *PoolWork

func NewPoolWork(pool *Pool) *PoolWork {
	return &PoolWork{
		ExtraNonce1:    pool.extraNonce1,
		ExtraNonce2Len: pool.extraNonce2Len,
		Difficulty:     pool.setDifficulty.Difficulty,
		JobId:          pool.notify.JobId,
		PrevHash:       pool.notify.PrevHash,
		CoinBase1:      pool.notify.CoinBase1,
		CoinBase2:      pool.notify.CoinBase2,
		MerkleBranches: pool.notify.MerkleBranches,
		Version:        pool.notify.Version,
		Nbits:          pool.notify.NBits,
		Ntime:          pool.notify.NTime,
		CleanJobs:      pool.notify.CleanJobs,
		Pool:           pool,
	}
}

func (ps *PoolWork) String() string {
	return fmt.Sprint("Work ", ps.JobId, " difficulty ", ps.Difficulty, " from ", ps.Pool)
}
