package workers

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	obc "github.com/kv-base-hack/base-binance-client"
	"github.com/kv-base-hack/base-token-rate/common"
	"github.com/kv-base-hack/base-token-rate/lib/rateprovider"
	"github.com/kv-base-hack/base-token-rate/storage/db"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"github.com/kv-base-hack/common/utils"
	"go.uber.org/zap"
)

const ratePricesKey = "dex_screener_prices"
const delayTime = time.Second / 3
const maxTokenPool = 30
const maxTokenNumber = 6
const maxBlockRange = 7200 * 30
const eth = "ethereum"
const sol = "solana"
const minTotalTradeIn24h = 100
const minTotalBuyIn24h = 10
const minLiquidity = 10000
const chunkPairWithUsdt = 80

type ChainData struct {
	lastStoredBlock int64
	tokenPools      map[string]int
}

type TokenPool struct {
	Address      string `json:"address"`
	NumberOfPool int    `json:"numberOfPool"`
}

type RateWorker struct {
	log                  *zap.SugaredLogger
	duration             time.Duration
	rateProvider         rateprovider.RateProvider
	inMemDB              inmem.Inmem
	db                   db.DB
	kaivestBinanceClient *obc.KaivestBinanceClient
	chainData            map[common.Chain]*ChainData
}

func NewRateWorker(log *zap.SugaredLogger, duration time.Duration,
	rateProvider rateprovider.RateProvider, inMemDB inmem.Inmem, db db.DB, kaivestBinanceClient *obc.KaivestBinanceClient) *RateWorker {
	return &RateWorker{
		log:                  log,
		duration:             duration,
		rateProvider:         rateProvider,
		inMemDB:              inMemDB,
		db:                   db,
		kaivestBinanceClient: kaivestBinanceClient,

		chainData: map[common.Chain]*ChainData{
			common.ChainBase: {
				lastStoredBlock: 0,
				tokenPools:      make(map[string]int),
			},
		},
	}
}

func (r *RateWorker) getCexMap(log *zap.SugaredLogger) map[string]float64 {
	ratesMap := map[string]float64{}
	// get rate from cex first
	pairWithUdst, err := r.kaivestBinanceClient.GetPairsWithUsdt()
	if err != nil {
		log.Errorw("error when get pair with usdt", "err", err)
		return map[string]float64{}
	}
	times := (len(pairWithUdst) + chunkPairWithUsdt - 1) / chunkPairWithUsdt
	for i := 0; i < times; i++ {
		bg := i * chunkPairWithUsdt
		end := bg + chunkPairWithUsdt
		if end > len(pairWithUdst) {
			end = len(pairWithUdst)
		}
		symbols := strings.Join(pairWithUdst[bg:end], ",")
		rates, err := r.kaivestBinanceClient.GetSpotBookTicker(symbols)
		if err != nil {
			log.Errorw("error when get book ticker for pairs", "symbols", symbols, "err", err)
			continue
		}
		for _, r := range rates {
			rate, err := strconv.ParseFloat(r.LastPrice, 64)
			if err != nil {
				log.Errorw("error when parse price", "r", r, "err", err)
				continue
			}
			ratesMap[r.Symbol] = rate
		}
	}

	return ratesMap
}

func (r *RateWorker) validNetwork(nw string) bool {
	network := []string{"ETH", "SOL"}
	for _, n := range network {
		if n == nw {
			return true
		}
	}
	return false
}

func (r *RateWorker) getNewAddresses(log *zap.SugaredLogger, chain common.Chain, lastStored int64, lastStoredBlockDb int64) []string {
	var tradeTable string
	var transferTable string
	if chain == common.ChainBase {
		tradeTable = db.BaseTradeLogs
		transferTable = db.BaseTransferLogs
	}
	newAddressTrades, err := r.db.GetUniqueTokenAddressByRangeForTrade(tradeTable, lastStored+1, lastStoredBlockDb)
	if err != nil {
		log.Errorw("error when new token address by range",
			"lastStored", lastStored, "lastStoredBlockDb", lastStoredBlockDb, "err", err)
		return []string{}
	}

	newAddressTransfer, err := r.db.GetUniqueTokenAddressByRangeForTransfer(transferTable, lastStored+1, lastStoredBlockDb)
	if err != nil {
		log.Errorw("error when new token address by range",
			"lastStored", lastStored, "lastStoredBlockDb", lastStoredBlockDb, "err", err)
		return []string{}
	}
	newAddress := append(newAddressTrades, newAddressTransfer...)
	return newAddress
}

func (r *RateWorker) updateTokenPoolFromBase(log *zap.SugaredLogger, existedOnCex map[string]bool) {
	lastEthStoredBlockDb, err := r.db.GetLastStoredBlock(db.BaseTradeLogs)
	if err != nil {
		log.Errorw("error when get last ethereum stored block in db", "err", err)
		return
	}
	lastStored := r.chainData[common.ChainBase].lastStoredBlock
	if lastStored < lastEthStoredBlockDb-maxBlockRange {
		lastStored = lastEthStoredBlockDb - maxBlockRange
	}

	log.Infow("set rate", "lastEthStoredBlockDb", lastEthStoredBlockDb, "lastStored", lastStored)

	// update new address for ethereum
	newAddress := r.getNewAddresses(log, common.ChainBase, lastStored+1, lastEthStoredBlockDb)
	for _, a := range newAddress {
		// get from cex, dont need to get from dex
		if _, exist := existedOnCex[a]; exist {
			continue
		}
		if _, exist := r.chainData[common.ChainBase].tokenPools[a]; !exist {
			// new pool for token
			r.chainData[common.ChainBase].tokenPools[a] = 0
		}
	}
	r.chainData[common.ChainBase].lastStoredBlock = lastEthStoredBlockDb
}

func (r *RateWorker) setRateToStorage() {
	log := r.log.With("ID", utils.RandomString(21))
	ratesMap := r.getCexMap(log)

	coins, err := r.kaivestBinanceClient.GetAllCoinInfo()
	if err != nil {
		log.Errorw("error when get all coins", "err", err)
	}

	tokens := []common.Token{}

	existedOnCex := map[string]bool{}

	for _, c := range coins {
		for _, n := range c.NetworkList {
			if !r.validNetwork(n.Network) {
				continue
			}
			tokenUsdt := n.Coin + "USDT"
			rate, exist := ratesMap[tokenUsdt]
			if !exist {
				continue
			}
			var chainID string
			if n.Network == "ETH" {
				chainID = eth
			} else if n.Network == "SOL" {
				chainID = sol
			}
			tokens = append(tokens, common.Token{
				UsdPrice:    rate,
				Address:     n.ContractAddress,
				Symbol:      n.Coin,
				ChainID:     chainID,
				SourcePrice: common.SourcePriceCex,
			})
			existedOnCex[strings.ToLower(n.ContractAddress)] = true
		}
	}
	log.Infow("finish get rate from cex", "tokens", tokens)
	r.updateTokenPoolFromBase(log, existedOnCex)
	tokenPool := []TokenPool{}
	for _, v := range r.chainData {
		for a, p := range v.tokenPools {
			tokenPool = append(tokenPool, TokenPool{
				Address:      a,
				NumberOfPool: p,
			})
		}
	}

	sort.Slice(tokenPool, func(i, j int) bool {
		return tokenPool[i].NumberOfPool < tokenPool[j].NumberOfPool
	})

	allPairs := []common.Pair{}
	totalPool := 0
	totalToken := 0
	addresses := []string{}
	for _, t := range tokenPool {
		if totalToken+1 > maxTokenNumber || totalPool+t.NumberOfPool > maxTokenPool {
			// call request
			tokens := strings.Join(addresses, ",")
			log.Infow("get rates for", "tokens", tokens)
			rates, err := r.rateProvider.GetPrices(tokens)
			if err != nil {
				log.Errorw("error when get rates", "err", err)
				continue
			}
			allPairs = append(allPairs, rates.Pairs...)
			time.Sleep(delayTime)

			totalToken = 1
			totalPool = t.NumberOfPool
			addresses = []string{t.Address}
			continue
		}
		totalToken += 1
		totalPool += t.NumberOfPool
		addresses = append(addresses, t.Address)
	}
	if totalToken > 0 {
		tokens := strings.Join(addresses, ",")
		log.Infow("get rates for", "tokens", tokens)
		rates, err := r.rateProvider.GetPrices(tokens)
		if err != nil {
			log.Errorw("error when get rates", "err", err)
		} else {
			allPairs = append(allPairs, rates.Pairs...)
		}
	}
	log.Infow("allPairs", "allPairs", allPairs)
	poolOfToken := map[string]int{}
	maxVolume := map[string]float64{}

	for _, p := range allPairs {
		if p.ChainID != eth && p.ChainID != sol {
			continue
		}
		// shouldn't get rate from stale pool
		if p.Txns.H24.Buys+p.Txns.H24.Sells <= minTotalTradeIn24h ||
			p.Txns.H24.Buys <= minTotalBuyIn24h || p.Liquidity.Usd < minLiquidity {
			continue
		}

		poolOfToken[p.BaseToken.Address]++
		if currentVolume, exist := maxVolume[p.BaseToken.Address]; exist {
			// choose the pool has max volume
			if currentVolume > p.Volume.H24 {
				continue
			}
		}
		maxVolume[p.BaseToken.Address] = p.Volume.H24
		tokens = append(tokens, common.Token{
			UsdPrice:    p.PriceUsd,
			Address:     p.BaseToken.Address,
			Symbol:      p.BaseToken.Symbol,
			ChainID:     p.ChainID,
			SourcePrice: common.SourcePriceDex,
			ImageUrl:    p.Info.ImageUrl,

			PriceChangeM5:  p.PriceChange.M5,
			PriceChangeH1:  p.PriceChange.H1,
			PriceChangeH6:  p.PriceChange.H6,
			PriceChangeH24: p.PriceChange.H24,
		})
	}

	for addr, value := range poolOfToken {
		r.chainData[common.ChainBase].tokenPools[addr] = value
	}

	log.Infow("tokens", "tokens", tokens)

	data, err := json.Marshal(tokens)
	if err != nil {
		r.log.Errorw("error when marshal data", "err", err)
		return
	}

	// no expire
	err = r.inMemDB.Set(ratePricesKey, data, 0)
	if err != nil {
		r.log.Errorw("error when set key", "key", ratePricesKey, "err", err)
	}
	r.log.Infow("finish set rates")
}

func (r *RateWorker) Run() error {
	log := r.log.With("worker", "rate_worker")
	log.Infow("start run rate worker")
	for {
		r.setRateToStorage()
		time.Sleep(r.duration)
	}
	return nil
}
