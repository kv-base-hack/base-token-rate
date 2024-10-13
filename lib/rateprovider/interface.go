package rateprovider

import "github.com/kv-base-hack/base-token-rate/common"

type RateProvider interface {
	GetPrices(tokenAddress string) (common.Pairs, error)
}
