package main

import (
	"time"

	"github.com/urfave/cli/v2"
)

const (
	tokenInfoWorkerDurationFlag = "token-info-worker-duration"
	cmcKeyFlag                  = "cmc-key"
	cmcUrlFlag                  = "cmc-url"
)

var tokenInfoFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    cmcKeyFlag,
		Usage:   "coinmarketcap key",
		EnvVars: []string{"CMC_KEY"},
	},
	&cli.StringFlag{
		Name:    cmcUrlFlag,
		Usage:   "coinmarketcap url",
		EnvVars: []string{"CMC_URL"},
	},
	&cli.DurationFlag{
		Name:    tokenInfoWorkerDurationFlag,
		Usage:   "get token info duration for worker",
		Value:   time.Hour * 12,
		EnvVars: []string{"TOKEN_INFO_WORKER_DURATION"},
	},
}

func NewTokenInfoFlags() (flags []cli.Flag) {
	return tokenInfoFlags
}
