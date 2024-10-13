package dexscreener

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kv-base-hack/base-token-rate/common"
	"github.com/kv-base-hack/common/utils"
	"go.uber.org/zap"
)

type DexScreener struct {
	log    *zap.SugaredLogger
	client *http.Client
	url    string
}

func NewDexScreener(log *zap.SugaredLogger, url string) *DexScreener {
	return &DexScreener{
		log:    log,
		client: &http.Client{},
		url:    url,
	}
}

func (d *DexScreener) GetPrices(tokenAddress string) (common.Pairs, error) {
	log := d.log.With("get_prices", utils.RandomString(22))
	path := d.url + fmt.Sprintf("/latest/dex/tokens/%s", tokenAddress)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Errorw("error when make request", "err", err)
		return common.Pairs{}, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		log.Errorw("Error sending request to server", "err", err)
		return common.Pairs{}, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorw("Error sending read resp body", "err", err)
		return common.Pairs{}, err
	}
	var dexScreenData common.Pairs
	if err := json.Unmarshal(respBody, &dexScreenData); err != nil {
		log.Errorw("Error sending parse to dex screener", "respBody", string(respBody), "err", err)
		return common.Pairs{}, err
	}

	log.Debugw("dexScreenData", "dexScreenData", dexScreenData)
	return dexScreenData, nil
}
