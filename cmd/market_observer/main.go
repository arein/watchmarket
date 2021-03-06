package main

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/trustwallet/blockatlas/pkg/logger"
	"github.com/trustwallet/watchmarket/internal"
	"github.com/trustwallet/watchmarket/market"
	"github.com/trustwallet/watchmarket/market/rate"
	rateCMC "github.com/trustwallet/watchmarket/market/rate/cmc"
	rateCoingecko "github.com/trustwallet/watchmarket/market/rate/coingecko"
	rateCompound "github.com/trustwallet/watchmarket/market/rate/compound"
	rateFixer "github.com/trustwallet/watchmarket/market/rate/fixer"
	"github.com/trustwallet/watchmarket/market/ticker"
	tickerCMC "github.com/trustwallet/watchmarket/market/ticker/cmc"
	tickerCoingecko "github.com/trustwallet/watchmarket/market/ticker/coingecko"
	tickerCompound "github.com/trustwallet/watchmarket/market/ticker/compound"
	tickerDEX "github.com/trustwallet/watchmarket/market/ticker/dex"
	"github.com/trustwallet/watchmarket/storage"
	"time"
)

const (
	defaultConfigPath              = "../../config.yml"
	gracefulShutdownTimeoutSeconds = 1
)

var (
	cache           *storage.Storage
	rateProviders   *rate.Providers
	tickerProviders *ticker.Providers
)

func init() {
	_, confPath := internal.ParseArgs("", defaultConfigPath)
	internal.InitConfig(confPath)
	logger.InitLogger()

	redisHost := viper.GetString("storage.redis")
	cache = internal.InitRedis(redisHost)

	rateProviders = &rate.Providers{
		// Add Market Quote Providers:
		0: rateCMC.InitRate(
			viper.GetString("market.cmc.api"),
			viper.GetString("market.cmc.api_key"),
			viper.GetString("market.cmc.map_url"),
			viper.GetString("market.rate_update_time"),
		),
		1: rateFixer.InitRate(
			viper.GetString("market.fixer.api"),
			viper.GetString("market.fixer.api_key"),
			viper.GetString("market.fixer.rate_update_time"),
		),
		2: rateCompound.InitRate(
			viper.GetString("market.compound.api"),
			viper.GetString("market.rate_update_time"),
		),
		3: rateCoingecko.InitRate(
			viper.GetString("market.coingecko.api"),
			viper.GetString("market.rate_update_time"),
		),
	}

	tickerProviders = &ticker.Providers{
		// Add Market Quote Providers:
		0: tickerCMC.InitMarket(
			viper.GetString("market.cmc.api"),
			viper.GetString("market.cmc.api_key"),
			viper.GetString("market.cmc.map_url"),
			viper.GetString("market.quote_update_time"),
		),
		1: tickerCompound.InitMarket(
			viper.GetString("market.compound.api"),
			viper.GetString("market.quote_update_time"),
		),
		2: tickerCoingecko.InitMarket(
			viper.GetString("market.coingecko.api"),
			viper.GetString("market.quote_update_time"),
		),
		3: tickerDEX.InitMarket(
			viper.GetString("market.dex.api"),
			viper.GetString("market.dex.quote_update_time"),
		),
	}
}

func main() {
	rateCron := market.InitRates(cache, rateProviders)
	defer gracefullyShutDown(rateCron)
	rateCron.Start()
	marketCron := market.InitTickers(cache, tickerProviders)
	defer gracefullyShutDown(marketCron)
	marketCron.Start()
	internal.WaitingForExitSignal()
	logger.Info("Waiting for all observer jobs to stop")
}

func gracefullyShutDown(job *cron.Cron) {
	c := job.Stop()
	select {
	case <-time.After(gracefulShutdownTimeoutSeconds * time.Second):
	case <-c.Done():
	}
}
