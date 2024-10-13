package main

import (
	"os"
	"sort"

	"github.com/joho/godotenv"
	obc "github.com/kv-base-hack/base-binance-client"
	"github.com/kv-base-hack/base-token-rate/lib/rateprovider/dexscreener"
	"github.com/kv-base-hack/base-token-rate/storage/db"
	"github.com/kv-base-hack/base-token-rate/workers"
	inmem "github.com/kv-base-hack/common/inmem_db"
	"github.com/kv-base-hack/common/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	app := cli.NewApp()
	app.Action = run
	app.Flags = append(app.Flags, logger.NewSentryFlags()...)
	app.Flags = append(app.Flags, NewPostgreSQLFlags()...)
	app.Flags = append(app.Flags, NewRateFlags()...)
	app.Flags = append(app.Flags, NewTokenInfoFlags()...)
	app.Flags = append(app.Flags, NewRedisFlags()...)
	sort.Sort(cli.FlagsByName(app.Flags))

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
func run(c *cli.Context) error {
	logger, flusher, err := logger.NewLogger(c)
	if err != nil {
		return err
	}
	defer flusher()
	zap.ReplaceGlobals(logger)
	log := logger.Sugar()
	log.Debugw("Starting application...")
	database, err := NewDBFromContext(c)
	if err != nil {
		log.Errorw("error when connect to database", "err", err)
		return err
	}
	pg := db.NewPostgres(database)

	redisHost := c.String(redisHostFlag)
	redisPort := c.String(redisPortFlag)
	redisPassword := c.String(redisPasswordFlag)
	redisDB := c.Int(redisDBFlag)
	redisAddr := redisHost + ":" + redisPort
	redis := inmem.NewRedisClient(redisAddr, redisPassword, redisDB)

	tokenInfo := workers.NewTokenInfoWorker(log, c.Duration(tokenInfoWorkerDurationFlag),
		c.String(cmcKeyFlag), c.String(cmcUrlFlag), redis)
	go tokenInfo.Run()

	kaivestBinance := obc.NewKaivestBinanceClient(c.String(kaivestBinanceUrlFlag))
	dexScreener := dexscreener.NewDexScreener(log, c.String(dexScreenerUrlFlag))
	rateWorkerDuration := workers.NewRateWorker(log, c.Duration(rateWorkerDuration), dexScreener, redis, pg, kaivestBinance)
	return rateWorkerDuration.Run()
}
