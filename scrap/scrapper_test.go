package scrap_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dk-open/crypto-zip/scrap/exchange/binance"
	"github.com/dk-open/crypto-zip/scrap/exchange/bitrue"
	"github.com/dk-open/crypto-zip/scrap/smart"
	"github.com/dk-open/crypto-zip/types"
	"testing"
	"time"
)

func TestScrapperBinance(t *testing.T) {
	ctx := context.Background()
	tm := time.Now()

	markets, err := binance.Markets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sMarkets := make(map[uint32]types.Market)

	for i, m := range markets {
		sMarkets[uint32(i)] = types.Market{
			Name:      m.Name,
			Precision: m.Precision,
		}
	}

	scrapper := smart.Scraper(sMarkets, binance.Prices)

	var buf bytes.Buffer
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	fmt.Println("Took", time.Since(tm).String(), buf.Len())

	time.Sleep(1000 * time.Millisecond)
	buf.Reset()
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	fmt.Println("Took", time.Since(tm).String(), buf.Len())

	time.Sleep(1000 * time.Millisecond)
	buf.Reset()
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	fmt.Println("Took", time.Since(tm).String(), buf.Len())

	time.Sleep(3000 * time.Millisecond)
	buf.Reset()
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	fmt.Println("Took", time.Since(tm).String(), buf.Len())

	time.Sleep(5000 * time.Millisecond)
	buf.Reset()
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	fmt.Println("Took", time.Since(tm).String(), buf.Len())
}

func TestScrapperBitrue(t *testing.T) {
	ctx := context.Background()
	tm := time.Now()

	markets, err := bitrue.Markets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sMarkets := make(map[uint32]types.Market)

	for i, m := range markets {
		sMarkets[uint32(i)] = types.Market{
			Name:      m.Name,
			Precision: m.Precision,
		}
	}
	fmt.Println(len(sMarkets))

	scrapper := smart.Scraper(sMarkets, bitrue.Prices)

	var buf bytes.Buffer
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	fmt.Println("Took", time.Since(tm).String(), buf.Len())

	time.Sleep(1000 * time.Millisecond)
	buf.Reset()
	tm = time.Now()
	if err = scrapper.Scrap(ctx, &buf); err != nil {
		t.Error(err)
	}
	//bbb := buf.Bytes()
	fmt.Println("Took", time.Since(tm).String(), buf.Len())
}
