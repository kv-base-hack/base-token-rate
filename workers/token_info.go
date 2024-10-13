package workers

import (
	"encoding/json"
	"time"

	"github.com/kv-base-hack/base-token-rate/common"
	"github.com/kv-base-hack/base-token-rate/lib/rateprovider/coinmarketcap"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"github.com/kv-base-hack/common/utils"
	"go.uber.org/zap"
)

const cmcTokenInfoKey = "cmc_token_info"

type TokenInfoWorker struct {
	log      *zap.SugaredLogger
	duration time.Duration
	cmc      *coinmarketcap.CoinMarketCap
	inMemDB  inmem.Inmem
}

func NewTokenInfoWorker(log *zap.SugaredLogger, duration time.Duration, key string, url string, inMemDB inmem.Inmem) *TokenInfoWorker {
	return &TokenInfoWorker{
		log:      log,
		duration: duration,
		cmc:      coinmarketcap.NewCoinMarketCap(log, key, url),
		inMemDB:  inMemDB,
	}
}

func (t *TokenInfoWorker) Run() {
	ticker := time.NewTicker(t.duration)
	for ; ; <-ticker.C {
		t.process()
	}
}

func (t *TokenInfoWorker) process() {
	start := int64(1)
	limit := int64(5000)
	tokenInfo := []common.RedisTokenInfo{}
	log := t.log.With("token_info", utils.RandomString(22))
	for {
		cmc, err := t.cmc.GetTokenInfo(start, limit)
		if err != nil {
			log.Errorw("error when get coinmarket cap token info", "err", err)
			return
		}
		log.Debugw("cmc info", "start", start, "limit", limit, "info", cmc.Status)
		for _, c := range cmc.Data {
			tokenInfo = append(tokenInfo, common.RedisTokenInfo{
				Name:                  c.Name,
				Symbol:                c.Symbol,
				CirculatingSupply:     c.CirculatingSupply,
				TotalSupply:           c.TotalSupply,
				MaxSupply:             c.MaxSupply,
				UsdPrice:              c.Quote.Usd.Price,
				MarketCap:             c.Quote.Usd.MarketCap,
				Tags:                  c.Tags,
				Volume24H:             c.Quote.Usd.Volume24H,
				FullyDilutedValuation: c.Quote.Usd.Price * c.MaxSupply,
				PercentChange1H:       c.Quote.Usd.PercentChange1H,
				PercentChange24H:      c.Quote.Usd.PercentChange24H,
				PercentChange7D:       c.Quote.Usd.PercentChange7D,
			})
		}
		if len(cmc.Data) != int(limit) {
			break
		}
		start = limit + 1
	}
	tokens := common.RedisTokens{
		UpdatedTime: time.Now().Unix(),
		Tokens:      tokenInfo,
	}
	data, err := json.Marshal(tokens)
	if err != nil {
		log.Errorw("error when marshal data", "err", err)
		return
	}

	// no expire
	err = t.inMemDB.Set(cmcTokenInfoKey, data, 0)
	if err != nil {
		log.Errorw("error when set key", "key", cmcTokenInfoKey, "err", err)
	}
	log.Infow("finish set token info")
}
