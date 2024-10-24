package smart

import (
	"bytes"
	"context"
	"github.com/dk-open/crypto-zip/compress"
	"github.com/dk-open/crypto-zip/scrap"
	"github.com/dk-open/crypto-zip/types"
	"sort"
)

type marketPrice struct {
	id        uint32
	precision float64
	bid       float64
	ask       float64
	updated   bool
}

type scrapper struct {
	markets []*marketPrice
	writer  scrap.IPriceWriter
	f       func(w scrap.IPriceWriter) error
}

func Scraper(markets map[uint32]types.Market, f func(w scrap.IPriceWriter) error) scrap.IScrapper {
	scrapMap := make(map[string]*marketPrice, len(markets))
	scrapMarkets := make([]*marketPrice, 0, len(markets))
	for id, m := range markets {
		mp := &marketPrice{id: id, precision: m.Precision}
		scrapMap[m.Name] = mp
		scrapMarkets = append(scrapMarkets, mp)
	}
	sort.Slice(scrapMarkets, func(i, j int) bool {
		return scrapMarkets[i].id < scrapMarkets[j].id
	})

	return &scrapper{
		f:       f,
		markets: scrapMarkets,
		writer:  PriceWriter(scrapMap),
	}
}

func (s *scrapper) Scrap(ctx context.Context, buf *bytes.Buffer) error {
	if err := s.f(s.writer); err != nil {
		return err
	}

	var index uint64
	for _, m := range s.markets {
		if m.updated {
			if err := compress.WriteVariant(buf, index); err != nil {
				return err
			}
			if err := compress.PackPrice(buf, uint64(m.bid*m.precision), uint64(m.ask*m.precision)-uint64(m.bid*m.precision)); err != nil {
				return err
			}
			index = 0
			continue
		}
		index++
	}
	return nil
}
