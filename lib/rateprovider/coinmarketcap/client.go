package coinmarketcap

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kv-base-hack/base-token-rate/common"
	"go.uber.org/zap"
)

type CoinMarketCap struct {
	log    *zap.SugaredLogger
	client *http.Client
	key    string
	url    string
}

func NewCoinMarketCap(log *zap.SugaredLogger, key string, url string) *CoinMarketCap {
	return &CoinMarketCap{
		log:    log,
		client: &http.Client{},
		key:    key,
		url:    url,
	}
}

func (c *CoinMarketCap) GetTokenInfo(start, limit int64) (common.CoinMarketCapTokenInfo, error) {
	path := c.url + "/v1/cryptocurrency/listings/latest"
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		c.log.Errorw("error when make request", "err", err)
		return common.CoinMarketCapTokenInfo{}, err
	}

	q := url.Values{}
	q.Add("start", strconv.FormatInt(start, 10))
	q.Add("limit", strconv.FormatInt(limit, 10))

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", c.key)
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Errorw("Error sending request to server", "err", err)
		return common.CoinMarketCapTokenInfo{}, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorw("Error sending read resp body", "err", err)
		return common.CoinMarketCapTokenInfo{}, err
	}
	var coinMarketCap common.CoinMarketCapTokenInfo
	if err := json.Unmarshal(respBody, &coinMarketCap); err != nil {
		c.log.Errorw("Error sending parse to coin market cap", "err", err)
		return common.CoinMarketCapTokenInfo{}, err
	}

	return coinMarketCap, nil
}
