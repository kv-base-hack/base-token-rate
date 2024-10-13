package common

import "time"

// enumer -type=Chain -linecomment -json=true -text=true -sql=true
type Chain uint64

const (
	ChainBase Chain = iota + 1 // base
)

// enumer -type=SourcePrice -linecomment -json=true -text=true -sql=true
type SourcePrice uint64

const (
	SourcePriceCex SourcePrice = iota + 1 // cex
	SourcePriceDex                        // dex
)

type Token struct {
	UsdPrice    float64     `json:"usdPrice"`
	Address     string      `json:"tokenAddress"`
	Symbol      string      `json:"symbol"`
	ChainID     string      `json:"chainId"`
	SourcePrice SourcePrice `json:"sourcePrice"`
	ImageUrl    string      `json:"imageUrl"`
	DexID       string      `json:"dexId"`
	Url         string      `json:"url"`

	PriceChangeM5  float64 `json:"priceChangeM5"`
	PriceChangeH1  float64 `json:"priceChangeH1"`
	PriceChangeH6  float64 `json:"priceChangeH6"`
	PriceChangeH24 float64 `json:"priceChangeH24"`
}

type TokenInfo struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Symbol            string    `json:"symbol"`
	Slug              string    `json:"slug"`
	CmcRank           int       `json:"cmc_rank,omitempty"`
	NumMarketPairs    int       `json:"num_market_pairs"`
	CirculatingSupply float64   `json:"circulating_supply"`
	TotalSupply       float64   `json:"total_supply"`
	MaxSupply         float64   `json:"max_supply"`
	LastUpdated       time.Time `json:"last_updated"`
	DateAdded         time.Time `json:"date_added"`
	Tags              []string  `json:"tags"`
	Platform          any       `json:"platform"`
	Quote             struct {
		Usd struct {
			Price            float64   `json:"price"`
			Volume24H        float64   `json:"volume_24h"`
			PercentChange1H  float64   `json:"percent_change_1h"`
			PercentChange24H float64   `json:"percent_change_24h"`
			PercentChange7D  float64   `json:"percent_change_7d"`
			MarketCap        float64   `json:"market_cap"`
			LastUpdated      time.Time `json:"last_updated"`
		} `json:"USD"`
	} `json:"quote,omitempty"`
}

type CoinMarketCapTokenInfo struct {
	Data   []TokenInfo `json:"data"`
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      int       `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
	} `json:"status"`
}

type CoinMarketCapPlatform struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Slug         string `json:"slug"`
	TokenAddress string `json:"token_address"`
}

type CoinMarketCapStatus struct {
	Timestamp    time.Time `json:"timestamp"`
	ErrorCode    int       `json:"error_code"`
	ErrorMessage string    `json:"error_message"`
	Elapsed      int       `json:"elapsed"`
	CreditCount  int       `json:"credit_count"`
}

type Pairs struct {
	Pairs []Pair `json:"pairs"`
}

type Pair struct {
	PriceUsd   float64   `json:"priceUsd,string"`
	BaseToken  PairToken `json:"baseToken"`
	QuoteToken PairToken `json:"quoteToken"`

	ChainID string `json:"chainId"`
	DexID   string `json:"dexId"`
	Url     string `json:"url"`
	Volume  struct {
		M5  float64 `json:"m5"`
		H1  float64 `json:"h1"`
		H6  float64 `json:"h6"`
		H24 float64 `json:"h24"`
	} `json:"volume"`
	PriceChange struct {
		M5  float64 `json:"m5"`
		H1  float64 `json:"h1"`
		H6  float64 `json:"h6"`
		H24 float64 `json:"h24"`
	} `json:"priceChange"`
	Txns struct {
		M5 struct {
			Buys  int64 `json:"buys"`
			Sells int64 `json:"sell"`
		} `json:"m5"`
		H1 struct {
			Buys  int64 `json:"buys"`
			Sells int64 `json:"sell"`
		} `json:"h1"`
		H6 struct {
			Buys  int64 `json:"buys"`
			Sells int64 `json:"sell"`
		} `json:"h6"`
		H24 struct {
			Buys  int64 `json:"buys"`
			Sells int64 `json:"sell"`
		} `json:"h24"`
	} `json:"txns"`
	Liquidity struct {
		Usd float64 `json:"usd"`
	} `json:"liquidity"`
	Info struct {
		ImageUrl string `json:"imageUrl"`
	} `json:"info"`
	Fdv float64 `json:"fdv"`
}

type PairToken struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type RedisTokens struct {
	UpdatedTime int64            `json:"updated_time"`
	Tokens      []RedisTokenInfo `json:"tokens"`
}

type RedisTokenInfo struct {
	Name                  string   `json:"name"`
	Symbol                string   `json:"symbol"`
	CirculatingSupply     float64  `json:"circulating_supply"`
	TotalSupply           float64  `json:"total_supply"`
	MaxSupply             float64  `json:"max_supply"`
	UsdPrice              float64  `json:"usd_price"`
	MarketCap             float64  `json:"market_cap"`
	Tags                  []string `json:"tags"`
	Volume24H             float64  `json:"volume_24h"`
	FullyDilutedValuation float64  `json:"fully_diluted_valuation"`
	PercentChange1H       float64  `json:"percent_change_1h"`
	PercentChange24H      float64  `json:"percent_change_24h"`
	PercentChange7D       float64  `json:"percent_change_7d"`
}
