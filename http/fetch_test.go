package http_test

import (
	"context"
	"fmt"
	"github.com/bcicen/jstream"
	"github.com/dk-open/crypto-zip/http"
	"github.com/dk-open/crypto-zip/types"
	"github.com/valyala/fastjson"
	"io"
	"testing"
)

type testSymbolModel struct {
	Symbol   string              `json:"symbol"`
	BidPrice types.StringToFloat `json:"bidPrice"`
	AskPrice types.StringToFloat `json:"askPrice"`
}

func TestFetcherWithDifferentEncodings(t *testing.T) {
	ctx := context.Background()
	targetUrl := "https://api.binance.com/api/v3/ticker/bookTicker"
	preCompressed, err := fetchAndCompress(targetUrl)
	if err != nil {
		t.Fatalf("Failed to fetch and compress data: %v", err)
	}

	client := createMockClient(preCompressed)

	fetcher := http.FetcherWithClient[[]testSymbolModel](client, "GET", targetUrl, http.WithCompression())
	var res []testSymbolModel
	if err = fetcher(ctx, &res); err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}
	if err = fetcher(ctx, &res); err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}
	if err = fetcher(ctx, &res); err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}
	if err = fetcher(ctx, &res); err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}

}

func TestFetcherFastIterate(t *testing.T) {
	targetUrl := "https://api.binance.com/api/v3/ticker/bookTicker"
	fetcher2 := http.Iterator[testSymbolModel]("GET", targetUrl, 1, http.WithCompression())
	if err := fetcher2(func(data testSymbolModel) error {
		fmt.Println(data)
		return nil
	}); err != nil {
		t.Fatalf("Failed to fetch data: %v", err)
	}
}

func BenchmarkFetcherGzip(b *testing.B) {
	ctx := context.Background()
	targetUrl := "https://www.bitrue.com/api/v1/ticker/24hr"
	preCompressed, err := fetchAndCompress(targetUrl)
	if err != nil {
		b.Fatalf("Failed to fetch and compress data: %v", err)
	}

	client := createMockClient(preCompressed)

	fetcherGzip := http.FetcherWithClient[[]testSymbolModel](client, "GET", targetUrl, http.WithHeader("Accept-Encoding", "gzip"))
	fetcherGzipFast := http.IteratorWithClient[testSymbolModel](client, "GET", targetUrl, 1, http.WithHeader("Accept-Encoding", "gzip"))
	fetcherBrFast := http.IteratorWithClient[testSymbolModel](client, "GET", targetUrl, 1, http.WithHeader("Accept-Encoding", "br"))
	fetcherPlainFast := http.IteratorWithClient[testSymbolModel](client, "GET", targetUrl, 1)
	fetcherBr := http.FetcherWithClient[[]testSymbolModel](client, "GET", targetUrl, http.WithHeader("Accept-Encoding", "br"))
	fetcherDeflate := http.FetcherWithClient[[]testSymbolModel](client, "GET", targetUrl, http.WithHeader("Accept-Encoding", "deflate"))
	fetcherPlain := http.FetcherWithClient[[]testSymbolModel](client, "GET", targetUrl)
	fetcherCustom := http.FetcherCustom(client, "GET", targetUrl, http.WithHeader("Accept-Encoding", "br"))

	b.Run("GZip", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var symbols []testSymbolModel
			err = fetcherGzip(ctx, &symbols)
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("Br", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var symbols []testSymbolModel
			err = fetcherBr(ctx, &symbols)
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("Deflate", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var symbols []testSymbolModel
			err = fetcherDeflate(ctx, &symbols)
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("plain", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var symbols []testSymbolModel
			err = fetcherPlain(ctx, &symbols)
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("CustomFastJson", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			symbols := make([]testSymbolModel, 0, 500)
			err = fetcherCustom(ctx, func(ctx context.Context, reader io.Reader) error {
				data, err := io.ReadAll(reader)
				_ = data

				p, err := fastjson.ParseBytes(data)
				for _, v := range p.GetArray() {
					symbols = append(symbols, testSymbolModel{
						Symbol: string(v.GetStringBytes("symbol")),
					})
				}
				return err
			})
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("CustomJstream", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			symbols := make([]testSymbolModel, 0, 500)
			_ = symbols
			err = fetcherCustom(ctx, func(ctx context.Context, reader io.Reader) error {
				decoder := jstream.NewDecoder(reader, 1000)
				_ = decoder

				for mv := range decoder.Stream() {
					_ = mv
				}
				return err
			})
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("GzipFast", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			//var symbols testSymbolModel
			err = fetcherGzipFast(func(data testSymbolModel) error {
				return nil
			})
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("BrFast", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			//var symbols testSymbolModel
			err = fetcherBrFast(func(data testSymbolModel) error {
				return nil
			})
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("PlainFast", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			//var symbols testSymbolModel
			err = fetcherPlainFast(func(data testSymbolModel) error {
				return nil
			})
			if err != nil {
				b.Fatalf("Fetcher failed: %v", err)
			}
		}
		b.ReportAllocs()
	})

}
