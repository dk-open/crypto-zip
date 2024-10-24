package smart

import "github.com/dk-open/crypto-zip/scrap"

type priceWriter struct {
	markets map[string]*marketPrice
}

func PriceWriter(marketsMap map[string]*marketPrice) scrap.IPriceWriter {
	return &priceWriter{markets: marketsMap}
}

func (w *priceWriter) Write(name string, bid, ask float64) error {
	if mp, ok := w.markets[name]; ok {
		if mp.bid != bid || mp.ask != ask {
			mp.updated = true
		} else {
			mp.updated = false
		}
		mp.bid = bid
		mp.ask = ask
	}
	return nil
}
