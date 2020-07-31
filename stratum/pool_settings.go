package stratum

type PoolSettings struct {
	ExtraNonce1    uint64
	ExtraNonce2Len int
	Difficulty     uint64
	JobId          string
	PrevHash       []byte
	CoinBase1      []byte
	CoinBase2      []byte
	MerkleBranch   [][]byte
	Version        []byte
	Nbits          []byte
	Ntime          []byte
	CleanJobs      bool
}

type PoolSettingsChan chan *PoolSettings

func NewPoolSettings(pool *Pool) *PoolSettings {
	return &PoolSettings{
		ExtraNonce1:    pool.extraNonce1,
		ExtraNonce2Len: pool.extraNonce2Len,
		Difficulty:     pool.difficulty,
		JobId:          pool.jobId,
		PrevHash:       pool.prevHash,
		CoinBase1:      pool.coinBase1,
		CoinBase2:      pool.coinBase2,
		MerkleBranch:   pool.merkleBranch,
		Version:        pool.version,
		Nbits:          pool.nbits,
		Ntime:          pool.ntime,
		CleanJobs:      pool.cleanJobs,
	}
}
