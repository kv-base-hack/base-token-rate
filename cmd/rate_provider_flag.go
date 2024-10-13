package main

import (
	"time"

	"github.com/urfave/cli/v2"
)

const (
	dexScreenerUrlFlag    = "dex-screener"
	rateWorkerDuration    = "rate-worker-duration"
	kaivestBinanceUrlFlag = "kaivest-binance-url"
)

var rateFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    dexScreenerUrlFlag,
		Usage:   "dex screener url",
		EnvVars: []string{"DEX_SCREENER"},
	},
	&cli.DurationFlag{
		Name:    rateWorkerDuration,
		Usage:   "rate worker duration",
		Value:   time.Minute,
		EnvVars: []string{"RATE_WORKER_DURATION"},
	},
	&cli.StringFlag{
		Name:    kaivestBinanceUrlFlag,
		Usage:   "kaivest binance url",
		EnvVars: []string{"KAIVEST_BINANCE_URL"},
	},
}

func NewRateFlags() (flags []cli.Flag) {
	return rateFlags
}
