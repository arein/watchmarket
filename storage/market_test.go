package storage

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trustwallet/watchmarket/mocks"
	"github.com/trustwallet/watchmarket/pkg/watchmarket"
	"testing"
	"time"
)

func TestSaveTickerWhenTickerDoesntExist(t *testing.T) {
	mockDb := &mocks.DB{}
	mockProviderList := &mocks.ProviderList{}
	mockTicker := &watchmarket.Ticker{
		Coin:       10,
		CoinName:   "myTestCoin",
		TokenId:    "myTestTokenId",
		CoinType:   "myTestCoinType",
		Price:      watchmarket.TickerPrice{},
		LastUpdate: time.Time{},
	}
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("**watchmarket.Ticker")).Return(watchmarket.ErrNotFound)
	mockDb.On("AddHM", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("*watchmarket.Ticker")).Return(nil)

	subject := &Storage{mockDb}

	err := subject.SaveTicker(mockTicker, mockProviderList)

	assert.Nil(t, err)
}

func TestSaveTickerWhenTickerDoesntExistAndDbFails(t *testing.T) {
	mockDb := &mocks.DB{}
	mockProviderList := &mocks.ProviderList{}
	mockTicker := &watchmarket.Ticker{
		Coin:       10,
		CoinName:   "myTestCoin",
		TokenId:    "myTestTokenId",
		CoinType:   "myTestCoinType",
		Price:      watchmarket.TickerPrice{},
		LastUpdate: time.Time{},
	}
	addHMErr := errors.New("boom")
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("**watchmarket.Ticker")).Return(watchmarket.ErrNotFound)
	mockDb.On("AddHM", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("*watchmarket.Ticker")).Return(addHMErr)

	subject := &Storage{mockDb}

	err := subject.SaveTicker(mockTicker, mockProviderList)

	assert.Equal(t, addHMErr, err)
}

func TestSaveTickerWhenTickerRetrievalFails(t *testing.T) {
	mockDb := &mocks.DB{}
	mockProviderList := &mocks.ProviderList{}
	mockTicker := &watchmarket.Ticker{
		Coin:       10,
		CoinName:   "myTestCoin",
		TokenId:    "myTestTokenId",
		CoinType:   "myTestCoinType",
		Price:      watchmarket.TickerPrice{},
		LastUpdate: time.Time{},
	}
	retrievalErr := errors.New("boom")
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("**watchmarket.Ticker")).Return(retrievalErr)

	subject := &Storage{mockDb}

	err := subject.SaveTicker(mockTicker, mockProviderList)

	assert.Equal(t, retrievalErr, err)
}

func TestSaveTickerWhenTickerExistsAndPriorityTooLow(t *testing.T) {
	mockDb := &mocks.DB{}
	mockProviderList := &mocks.ProviderList{}
	mockProviderList.On("GetPriority", "myNewTestProvider").Return(10)
	mockProviderList.On("GetPriority", "myOldTestProvider").Return(0)
	mockTicker := &watchmarket.Ticker{
		CoinName:   "myTestCoin",
		TokenId:    "myTestTokenId",
		Price:      watchmarket.TickerPrice{
			Provider:  "myNewTestProvider",
		},
		LastUpdate: time.Now(),
	}
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.MatchedBy(func(value interface{}) bool {
		*value.(**watchmarket.Ticker) = &watchmarket.Ticker{
			CoinName:   "myTestCoin",
			TokenId:    "myTestTokenId",
			Price:      watchmarket.TickerPrice{
				Provider:  "myOldTestProvider",
			},
			LastUpdate: time.Now(),
		}
		return true
	})).Return(nil)

	subject := &Storage{mockDb}

	err := subject.SaveTicker(mockTicker, mockProviderList)

	assert.Contains(t, err.Error(), "less priority")
}

func TestSaveTickerWhenTickerExistsAndOutdated(t *testing.T) {
	mockDb := &mocks.DB{}
	mockProviderList := &mocks.ProviderList{}
	mockProviderList.On("GetPriority", "myNewTestProvider").Return(0)
	mockProviderList.On("GetPriority", "myOldTestProvider").Return(10)
	mockTicker := &watchmarket.Ticker{
		CoinName:   "myTestCoin",
		TokenId:    "myTestTokenId",
		Price:      watchmarket.TickerPrice{
			Provider:  "myNewTestProvider",
		},
		LastUpdate: time.Now(),
	}
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.MatchedBy(func(value interface{}) bool {
		*value.(**watchmarket.Ticker) = &watchmarket.Ticker{
			CoinName:   "myTestCoin",
			TokenId:    "myTestTokenId",
			Price:      watchmarket.TickerPrice{
				Provider:  "myOldTestProvider",
			},
			LastUpdate: time.Now(),
		}
		return true
	})).Return(nil)

	subject := &Storage{mockDb}

	err := subject.SaveTicker(mockTicker, mockProviderList)

	assert.Contains(t, err.Error(), "outdated")
}

func TestGetTickerNominal(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("**watchmarket.Ticker")).Return(nil)

	subject := &Storage{mockDb}

	_, err := subject.GetTicker("myTestCoin", "myTestTokenId")

	assert.Nil(t, err)
}

func TestGetTickerWithoutToken(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN", mock.AnythingOfType("**watchmarket.Ticker")).Return(nil)

	subject := &Storage{mockDb}

	_, err := subject.GetTicker("myTestCoin", "")

	assert.Nil(t, err)
}

func TestGetTickerError(t *testing.T) {
	dbErr := errors.New("boom")
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_QUOTES", "MYTESTCOIN_MYTESTTOKENID", mock.AnythingOfType("**watchmarket.Ticker")).Return(dbErr)

	subject := &Storage{mockDb}

	_, err := subject.GetTicker("myTestCoin", "myTestTokenId")

	assert.Equal(t, dbErr, err)
}

func TestGetRateNominal(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("**watchmarket.Rate")).Return(nil)

	subject := &Storage{mockDb}

	_, err := subject.GetRate("myTestCurrency")

	assert.Nil(t, err)
}

func TestGetRateError(t *testing.T) {
	dbErr := errors.New("boom")
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("**watchmarket.Rate")).Return(dbErr)

	subject := &Storage{mockDb}

	_, err := subject.GetRate("myTestCurrency")

	assert.Equal(t, dbErr, err)
}

func TestSaveRatesNoExistingRate(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("**watchmarket.Rate")).Return(watchmarket.ErrNotFound)
	mockDb.On("AddHM", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("*watchmarket.Rate")).Return(nil)
	mockRates := watchmarket.Rates{}
	mockRates = append(mockRates, watchmarket.Rate{Currency: "myTestCurrency",})
	mockProviderList := &mocks.ProviderList{}

	subject := &Storage{mockDb}

	subject.SaveRates(mockRates, mockProviderList)

	//  TBD Assertions
}

func TestSaveRatesExistingRate(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.MatchedBy(func(value interface{}) bool {
		*value.(**watchmarket.Rate) = &watchmarket.Rate{
			Currency:         "myTestCurrency",
			Provider:         "myOldTestProvider",
			Timestamp:        0,
		}
		return true
	})).Return(nil)
	mockDb.On("AddHM", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("*watchmarket.Rate")).Return(nil)
	mockRates := watchmarket.Rates{}
	mockRates = append(mockRates, watchmarket.Rate{
		Currency:         "myTestCurrency",
		Provider:         "myNewTestProvider",
		Timestamp:        10,
	})
	mockProviderList := &mocks.ProviderList{}
	mockProviderList.On("GetPriority", "myNewTestProvider").Return(0)
	mockProviderList.On("GetPriority", "myOldTestProvider").Return(10)

	subject := &Storage{mockDb}

	subject.SaveRates(mockRates, mockProviderList)

	//  TBD Assertions
}

func TestSaveRatesExistingRateNewRateLowPriority(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.MatchedBy(func(value interface{}) bool {
		*value.(**watchmarket.Rate) = &watchmarket.Rate{
			Currency:         "myTestCurrency",
			Provider:         "myOldTestProvider",
			Timestamp:        0,
		}
		return true
	})).Return(nil)
	mockDb.On("AddHM", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("*watchmarket.Rate")).Return(nil)
	mockRates := watchmarket.Rates{}
	mockRates = append(mockRates, watchmarket.Rate{
		Currency:         "myTestCurrency",
		Provider:         "myNewTestProvider",
		Timestamp:        10,
	})
	mockProviderList := &mocks.ProviderList{}
	mockProviderList.On("GetPriority", "myNewTestProvider").Return(10)
	mockProviderList.On("GetPriority", "myOldTestProvider").Return(0)

	subject := &Storage{mockDb}

	subject.SaveRates(mockRates, mockProviderList)

	//  TBD Assertions
}

func TestSaveRatesExistingRateNewRateOutdated(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.MatchedBy(func(value interface{}) bool {
		*value.(**watchmarket.Rate) = &watchmarket.Rate{
			Currency:         "myTestCurrency",
			Provider:         "myOldTestProvider",
			Timestamp:        10,
		}
		return true
	})).Return(nil)
	mockDb.On("AddHM", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("*watchmarket.Rate")).Return(nil)
	mockRates := watchmarket.Rates{}
	mockRates = append(mockRates, watchmarket.Rate{
		Currency:         "myTestCurrency",
		Provider:         "myNewTestProvider",
		Timestamp:        0,
	})
	mockProviderList := &mocks.ProviderList{}
	mockProviderList.On("GetPriority", "myNewTestProvider").Return(0)
	mockProviderList.On("GetPriority", "myOldTestProvider").Return(0)

	subject := &Storage{mockDb}

	subject.SaveRates(mockRates, mockProviderList)

	//  TBD Assertions
}

func TestSaveRatesDbSaveFails(t *testing.T) {
	mockDb := &mocks.DB{}
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("**watchmarket.Rate")).Return(watchmarket.ErrNotFound)
	dbSaveFailure := errors.New("boom")
	mockDb.On("AddHM", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("*watchmarket.Rate")).Return(dbSaveFailure)
	mockRates := watchmarket.Rates{}
	mockRates = append(mockRates, watchmarket.Rate{Currency: "myTestCurrency",})
	mockProviderList := &mocks.ProviderList{}

	subject := &Storage{mockDb}

	subject.SaveRates(mockRates, mockProviderList)

	//  TBD Assertions
}
func TestSaveRatesDbRetrievalFails(t *testing.T) {
	mockDb := &mocks.DB{}
	dbRetrievalFailure := errors.New("boom")
	mockDb.On("GetHMValue", "ATLAS_MARKET_RATES", "myTestCurrency", mock.AnythingOfType("**watchmarket.Rate")).Return(dbRetrievalFailure)
	mockRates := watchmarket.Rates{}
	mockRates = append(mockRates, watchmarket.Rate{Currency: "myTestCurrency",})
	mockProviderList := &mocks.ProviderList{}

	subject := &Storage{mockDb}

	subject.SaveRates(mockRates, mockProviderList)

	//  TBD Assertions
}