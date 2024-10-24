package compress_test

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/dk-open/crypto-zip/compress"
	"math/rand"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
)

func TestPriceEncoding(t *testing.T) {
	// Create a sample price instance
	p := compress.Price(1234567890, 987654321)
	// Encode using Bytes method
	bytesOutput := p.Bytes()

	// Encode using Write method
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("Write method returned error: %v", err)
	}
	writeOutput := buf.Bytes()

	// Compare the outputs
	if !bytes.Equal(bytesOutput, writeOutput) {
		t.Errorf("Outputs differ between Bytes and Write methods.\nBytes Output: %x\nWrite Output: %x", bytesOutput, writeOutput)
	}
	fmt.Printf("Bytes Output: %x\n", bytesOutput)
	fmt.Printf("Write Output: %x\n", writeOutput)

	fmt.Println()
	// Decode the outputs to verify correctness
	//verifyEncoding(t, bytesOutput, p.bid, p.askDiff, "Bytes Output")
	//verifyEncoding(t, writeOutput, p.bid, p.askDiff, "Write Output")
}

func TestPricesEncoding(t *testing.T) {
	numPrices := 500
	rng := rand.New(rand.NewSource(1))
	prices := make([]compress.IPrice, numPrices)
	for i := 0; i < numPrices; i++ {
		bid := rng.Uint64() % 100_000_000
		askDiffMax := bid / 10
		if askDiffMax == 0 {
			askDiffMax = 1
		}
		askDiff := rng.Uint64() % askDiffMax
		prices[i] = compress.Price(bid, askDiff)
	}
	// Process all 500 prices using Bytes method
	var data []byte
	for _, p := range prices {
		data = append(data, p.Bytes()...)
	}
	fmt.Printf("Bytes Output: %d %x\n", len(data), data)

	var buf bytes.Buffer
	for _, p := range prices {
		if err := p.Write(&buf); err != nil {
			t.Fatalf("Write method returned error: %v", err)
		}
	}
	data2 := buf.Bytes()
	fmt.Printf("Bytes Output: %d %x\n", len(data2), data2)

}

func BenchmarkCompress500Prices(b *testing.B) {
	// Prepare 500 random prices
	numPrices := 500
	prices := make([]compress.IPrice, numPrices)
	rng := rand.New(rand.NewSource(1))

	for i := 0; i < numPrices; i++ {
		bid := rng.Uint64() % 100_000_000
		askDiffMax := bid / 10
		if askDiffMax == 0 {
			askDiffMax = 1
		}
		askDiff := rng.Uint64() % askDiffMax
		prices[i] = compress.Price(bid, askDiff)
	}

	b.ResetTimer()

	// Benchmark the Bytes method
	b.Run("BytesMethod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Process all 500 prices using Bytes method
			var data []byte
			for _, p := range prices {
				data = append(data, p.Bytes()...)
			}
			// Use data if needed
			_ = data
		}
		b.ReportAllocs()
	})

	// Benchmark the Write method
	b.Run("WriteMethod", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			buf.Reset()
			// Process all 500 prices using Write method
			for _, p := range prices {
				if err := p.Write(&buf); err != nil {
					b.Fatalf("Write method returned error: %v", err)
				}
			}
			// Use buf.Bytes() if needed
			data := buf.Bytes()
			_ = data
		}
		b.ReportAllocs()
	})
}

func BenchmarkCompressionAlgorithms(b *testing.B) {
	// Generate random data buffer
	numPrices := 500
	prices := make([]compress.IPrice, numPrices)
	rng := rand.New(rand.NewSource(1))

	for i := 0; i < numPrices; i++ {
		bid := rng.Uint64() % 100_000_000
		askDiffMax := bid / 10
		if askDiffMax == 0 {
			askDiffMax = 1
		}
		askDiff := rng.Uint64() % askDiffMax
		prices[i] = compress.Price(bid, askDiff)
	}
	var bf bytes.Buffer
	for _, p := range prices {
		if err := p.Write(&bf); err != nil {
			b.Fatalf("Write method returned error: %v", err)
		}
	}
	// Use buf.Bytes() if needed
	data := bf.Bytes()
	fmt.Printf("Original: %d\n", len(data))

	var compressedData []byte

	// Benchmark gzip
	b.Run("Gzip", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			_, err := gzWriter.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			gzWriter.Close()
			compressedData = buf.Bytes()
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
	})

	// Benchmark zlib
	b.Run("Zlib", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			zlibWriter := zlib.NewWriter(&buf)
			_, err := zlibWriter.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			zlibWriter.Close()
			compressedData = buf.Bytes()
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
	})

	// Benchmark flate
	b.Run("Flate", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			flateWriter, err := flate.NewWriter(&buf, flate.DefaultCompression)
			if err != nil {
				b.Fatal(err)
			}
			_, err = flateWriter.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			flateWriter.Close()
			compressedData = buf.Bytes()
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
	})

	// Benchmark Snappy
	b.Run("Snappy", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			compressedData = snappy.Encode(nil, data)
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
	})

	// Benchmark LZ4
	b.Run("LZ4", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			lz4Writer := lz4.NewWriter(&buf)

			_, err := lz4Writer.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			lz4Writer.Close()
			compressedData = buf.Bytes()
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
	})

	b.Run("Zstd", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		encoder, err := zstd.NewWriter(nil)
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			compressedData = encoder.EncodeAll(data, nil)
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
		encoder.Close()
	})

	b.Run("Brotli", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			brotliWriter := brotli.NewWriter(&buf)
			_, err := brotliWriter.Write(data)
			if err != nil {
				b.Fatal(err)
			}
			brotliWriter.Close()
			compressedData = buf.Bytes()
		}
		b.ReportMetric(float64(len(compressedData)), "compressed_bytes")
		b.ReportMetric(float64(len(compressedData))/float64(len(data)), "compression_ratio")
	})
}
