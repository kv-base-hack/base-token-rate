package moralis

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/kv-base-hack/base-token-rate/common"
)

type MoralisClient struct {
	chain    string
	url      string
	keys     []string
	keyIndex int
}

func NewMoralisClient(chain string, url string, key string) *MoralisClient {
	keys := strings.Split(key, ",")

	return &MoralisClient{
		chain:    chain,
		url:      url,
		keys:     keys,
		keyIndex: 0,
	}
}

func (c *MoralisClient) GetPrices(tokenAddress []string) ([]common.Token, error) {
	tokens := Tokens{}
	for _, t := range tokenAddress {
		tokens.Tokens = append(tokens.Tokens, Token{
			TokenAddress: t,
		})
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		return nil, err
	}
	payload := bytes.NewBuffer(data)

	url := c.url + "/erc20/prices?chain=" + c.chain
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", c.keys[c.keyIndex])
	c.keyIndex = (c.keyIndex + 1) % len(c.keys)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tokenPrices []common.Token
	if err := json.Unmarshal(body, &tokenPrices); err != nil {
		return nil, err
	}
	return tokenPrices, nil
}
