package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // sql driver name: "postgres"
	"go.uber.org/zap"
)

const (
	BaseTradeLogs    = "base_trade_logs"
	BaseTransferLogs = "base_transfer_logs"
)

type Postgres struct {
	db *sqlx.DB
	l  *zap.SugaredLogger
}

func NewPostgres(db *sqlx.DB) *Postgres {
	return &Postgres{
		db: db,
		l:  zap.S(),
	}
}

func (pg *Postgres) GetLastStoredBlock(table string) (int64, error) {
	query, _, err := sq.
		Select("MAX(block_number) as block_number").
		From(table).ToSql()
	if err != nil {
		return 0, err
	}
	var result int64
	err = pg.db.Get(&result, query)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (pg *Postgres) GetUniqueTokenAddressByRangeForTrade(table string, from, to int64) ([]string, error) {
	firstSelect := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("token_in_address").From(table).
		Where(sq.And{sq.GtOrEq{"block_number": from}, sq.LtOrEq{"block_number": to}})
	secondSelect := sq.Select("token_out_address").From(table).
		Where(sq.And{sq.GtOrEq{"block_number": from}, sq.LtOrEq{"block_number": to}})

	sql, args, _ := secondSelect.ToSql()
	unionSelect := firstSelect.Suffix("UNION "+sql, args...)

	q, p, err := unionSelect.ToSql()
	if err != nil {
		return nil, err
	}
	var result []string
	err = pg.db.Select(&result, q, p...)

	return result, err
}

func (pg *Postgres) GetUniqueTokenAddressByRangeForTransfer(table string, from, to int64) ([]string, error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("distinct(token_address)").From(table).
		Where(sq.And{sq.GtOrEq{"block_number": from}, sq.LtOrEq{"block_number": to}})

	sql, args, _ := query.ToSql()
	var result []string
	err := pg.db.Select(&result, sql, args...)

	return result, err
}
