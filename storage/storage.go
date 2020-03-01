package storage

import (
	"github.com/trustwallet/watchmarket/pkg/watchmarket"
)

type Storage struct {
	DB
}

func New() *Storage {
	s := new(Storage)
	return s
}

type DB interface {
	GetHMValue(entity, key string, value interface{}) error
	AddHM(entity, key string, value interface{}) error
	Init(host string) error
}

type Market interface {
	SaveTicker(coin *watchmarket.Ticker, pl ProviderList) error
	GetTicker(coin, token string) (*watchmarket.Ticker, error)
	SaveRates(rates watchmarket.Rates, pl ProviderList)
	GetRate(currency string) (*watchmarket.Rate, error)
}
