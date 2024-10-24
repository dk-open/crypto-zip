package binance

import (
	"context"
	"github.com/dk-open/crypto-zip/scrap"
	"github.com/dk-open/crypto-zip/scrap/exchange"
	"github.com/dk-open/crypto-zip/tools/http"
	"math"
	"strconv"
)

type bookPrices struct {
	Symbol string  `json:"symbol"`
	Bid    float64 `json:"bidPrice"`
	Ask    float64 `json:"askPrice"`
}

type exchangeInfo struct {
	ServerTime int64 `json:"serverTime"`
	Symbols    []struct {
		Market              string                   `json:"symbol"`
		Status              string                   `json:"status"`
		BaseAsset           string                   `json:"baseAsset"`
		BaseAssetPrecision  int                      `json:"baseAssetPrecision"`
		QuoteAsset          string                   `json:"quoteAsset"`
		QuoteAssetPrecision int                      `json:"quoteAssetPrecision"`
		OrderTypes          []string                 `json:"orderTypes"`
		Filters             []map[string]interface{} `json:"filters"`
		Permissions         []string                 `json:"permissions"`
	} `json:"symbols"`
}

var marketsFetcher = http.Fetcher[exchangeInfo]("GET", "https://api.binance.com/api/v1/exchangeInfo")
var priceFetcher = http.Iterator[bookPrices]("GET", "https://api.binance.com/api/v3/ticker/bookTicker", 1)

func Prices(buf scrap.IPriceWriter) (err error) {
	return priceFetcher(func(v bookPrices) error {
		if v.Bid > 0. && v.Ask > 0. {
			if err = buf.Write(v.Symbol, v.Bid, v.Ask); err != nil {
				return err
			}
		}
		return nil
	})
}

func Markets(ctx context.Context) (res []exchange.Market, err error) {
	var info exchangeInfo
	if err = marketsFetcher(ctx, &info); err != nil {
		return nil, err
	}

	for _, pair := range info.Symbols {
		if pair.Status == "TRADING" {
			res = append(res, exchange.Market{
				Name:      pair.Market,
				Base:      pair.BaseAsset,
				Quote:     pair.QuoteAsset,
				Precision: math.Pow(10, float64(getPricePrecision(pair.Filters))),
			})
		}
	}
	return
}

func getPricePrecision(filters []map[string]interface{}) (res uint8) {
	for _, v := range filters {
		if v["filterType"] == "PRICE_FILTER" {
			if resPrec, ok := v["tickSize"]; ok {
				if precStr, sok := resPrec.(string); sok {
					f, _ := strconv.ParseFloat(precStr, 64)
					return DigitsAfterDot(f)
				}
			}
		}
	}
	return
}

func DigitsAfterDot(f float64) uint8 {
	// Subtract integer part to get the fraction
	fraction := f - math.Floor(f)

	var count uint8 = 0
	// While fraction is not an integer
	for fraction != math.Floor(fraction) {
		fraction *= 10
		count++
	}

	return count
}
