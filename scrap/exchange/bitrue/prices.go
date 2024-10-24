package bitrue

import (
	"context"
	"fmt"
	"github.com/dk-open/crypto-zip/http"
	"github.com/dk-open/crypto-zip/scrap"
	"github.com/dk-open/crypto-zip/scrap/exchange"
	"math"
)

type bookPrices struct {
	Symbol string  `json:"symbol"`
	Bid    float64 `json:"bidPrice"`
	Ask    float64 `json:"askPrice"`
}

type marketsData struct {
	Symbols []struct {
		Symbol    string `json:"symbol"`
		Status    string `json:"status"`
		Sell      string `json:"baseAsset"`
		Buy       string `json:"quoteAsset"`
		Precision uint8  `json:"quotePrecision"`
	} `json:"symbols"`
}

var marketsFetcher = http.Fetcher[marketsData]("GET", "https://openapi.bitrue.com/api/v1/exchangeInfo")
var priceFetcher = http.Iterator[bookPrices]("GET", "https://bitrue.com/api/v1/ticker/24hr", 1)

func Prices(buf scrap.IPriceWriter) (err error) {
	fmt.Println("Get Prices")
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
	var info marketsData
	if err = marketsFetcher(ctx, &info); err != nil {
		return nil, err
	}

	for _, pair := range info.Symbols {
		if pair.Status == "TRADING" {
			res = append(res, exchange.Market{
				Name:      pair.Symbol,
				Base:      pair.Sell,
				Quote:     pair.Buy,
				Precision: math.Pow(10, float64(pair.Precision)),
			})
		}
	}
	return
}
