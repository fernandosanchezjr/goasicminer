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
		Difficulty:     pool.difficulty,
		JobId:          pool.jobId,
		PrevHash:       pool.prevHash,
		CoinBase1:      pool.coinBase1,
		CoinBase2:      pool.coinBase2,
		MerkleBranches: pool.merkleBranches,
		Version:        pool.version,
		Nbits:          pool.nbits,
		Ntime:          pool.ntime,
		CleanJobs:      pool.cleanJobs,
		Pool:           pool,
	}
}

func (ps *PoolWork) String() string {
	return fmt.Sprint("Job ", ps.JobId, " from ", ps.Pool)
}
