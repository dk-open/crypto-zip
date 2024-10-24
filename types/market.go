package types

type Market struct {
	Name      string
	Precision float64
}

type Price [2]float64

type AssetID uint32
type MarketID uint64
type ExchangeID uint16
type ExchangeMarketID uint64

const AssetBits int = 24
const MarketBits int = 2 * AssetBits

func (a AssetID) ID() uint32 {
	return uint32(a)
}

func (e ExchangeID) ID() uint16 {
	return uint16(e)
}

type AssetMap map[string]AssetID

func (am AssetMap) Add(id AssetID, name string) {
	am[name] = id
}

func NewMarket(sell, buy AssetID) (res MarketID) {
	return MarketID(sell)<<AssetBits | MarketID(buy)
}

func (m MarketID) ID() uint32 {
	return uint32(m)
}

var maskAsset = AssetID(0xFFFFFF)

func (m MarketID) Revert() (res MarketID) {
	return (m>>AssetBits)&0xFFFFFF | (m&0xFFFFFF)<<AssetBits
}

func (m MarketID) Sell() AssetID {
	return AssetID(m>>AssetBits) & maskAsset
}

func (m MarketID) Buy() AssetID {
	return AssetID(m) & maskAsset
}

type IMarketAdapter interface {
	MarketID(sell, buy string) (MarketID, bool)
}

func (am AssetMap) MarketID(sell, buy string) (MarketID, bool) {
	id1, ok1 := am[sell]
	if !ok1 {
		return 0, false
	}
	id2, ok2 := am[buy]
	if !ok2 {
		return 0, false
	}
	return NewMarket(id1, id2), true
}

func ExchangeMarketByAssets(exchange uint16, sell, buy AssetID) ExchangeMarketID {
	return ExchangeMarketID(exchange)<<MarketBits | ExchangeMarketID(sell)<<AssetBits | ExchangeMarketID(buy)
}

func ExchangeMarket(exchange uint16, id uint32) ExchangeMarketID {
	return ExchangeMarketID(exchange)<<MarketBits | ExchangeMarketID(id)
}

func (m ExchangeMarketID) Exchange() ExchangeID {
	return ExchangeID(m >> MarketBits)
}

func (m ExchangeMarketID) Market() MarketID {
	return MarketID(m & 0xFFFFFFFFFFFF)
}

func (m ExchangeMarketID) Sell() AssetID {
	return m.Market().Sell()
}

func (m ExchangeMarketID) Buy() AssetID {
	return m.Market().Buy()
}

func (m ExchangeMarketID) Revert() ExchangeMarketID {
	return ExchangeMarket(m.Exchange().ID(), m.Market().Revert().ID())
}

func (m ExchangeMarketID) ID() uint64 {
	return uint64(m)
}
