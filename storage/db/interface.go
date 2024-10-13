package db

type DB interface {
	GetLastStoredBlock(table string) (int64, error)
	GetUniqueTokenAddressByRangeForTrade(table string, from, to int64) ([]string, error)
	GetUniqueTokenAddressByRangeForTransfer(table string, from, to int64) ([]string, error)
}
