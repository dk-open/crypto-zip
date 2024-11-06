package http_test

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// fetcher/fetcher_test.go (continued)

func createMockClient(preCompressed *PreCompressedData) *http.Client {
	mockRT := &MockRoundTripper{
		PreCompressed: preCompressed,
	}
	return &http.Client{
		Transport: mockRT,
	}
}

// MockRoundTripper is a custom RoundTripper for mocking HTTP responses.
type MockRoundTripper struct {
	// Map from Content-Encoding to file path
	PreCompressed *PreCompressedData
}

// RoundTrip implements the RoundTripper interface.
func (mrt *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Determine which encoding to use based on the request's Accept-Encoding header
	acceptEnc := req.Header.Get("Accept-Encoding")
	var encoding string
	encodings := []string{"identity"}
	if bytes.Contains([]byte(acceptEnc), []byte("br")) {
		encodings = append(encodings, "br")
	}
	if bytes.Contains([]byte(acceptEnc), []byte("gzip")) {
		encodings = append(encodings, "gzip")
	}
	if bytes.Contains([]byte(acceptEnc), []byte("deflate")) {
		encodings = append(encodings, "deflate")
	}
	if bytes.Contains([]byte(acceptEnc), []byte("zstd")) {
		encodings = append(encodings, "zstd")
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	encoding = encodings[rand.Int63n(int64(len(encodings)))]

	var data []byte
	var contentEncoding string

	switch encoding {
	case "br":
		data = mrt.PreCompressed.Brotli
		contentEncoding = "br"
	case "gzip":
		data = mrt.PreCompressed.Gzip
		contentEncoding = "gzip"
	case "zstd":
		data = mrt.PreCompressed.Zstd
	case "deflate":
		data = mrt.PreCompressed.Deflate
		contentEncoding = "deflate"
	default:
		data = mrt.PreCompressed.Identity
		contentEncoding = "identity"
	}

	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
	}

	if contentEncoding != "identity" {
		resp.Header.Set("Content-Encoding", contentEncoding)
	}

	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// PreCompressedData holds compressed data in different encodings.
type PreCompressedData struct {
	Gzip     []byte
	Deflate  []byte
	Brotli   []byte
	Zstd     []byte
	Identity []byte
}

// fetchAndCompress fetches data from Binance and compresses it.
func fetchAndCompress(url string) (*PreCompressedData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid response. Error code %d", resp.StatusCode)
	}

	identityData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Compress using gzip
	var gzipBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuf)
	_, err = gzipWriter.Write(identityData)
	if err != nil {
		return nil, fmt.Errorf("Failed to write gzip data: %v", err)
	}
	gzipWriter.Close()

	// Compress using deflate
	var deflateBuf bytes.Buffer
	deflateWriter := zlib.NewWriter(&deflateBuf)
	_, err = deflateWriter.Write(identityData)
	if err != nil {
		return nil, fmt.Errorf("Failed to write deflate data: %v", err)
	}
	deflateWriter.Close()

	// Compress using deflate
	var zstdBuf bytes.Buffer
	zstdWriter, err := zstd.NewWriter(&zstdBuf)
	if err != nil {
		return nil, fmt.Errorf("Failed to write zstd data: %v", err)
	}
	_, err = zstdWriter.Write(identityData)
	if err != nil {
		return nil, fmt.Errorf("Failed to write zstd data: %v", err)
	}
	zstdWriter.Flush()
	zstdWriter.Close()

	// Compress using brotli
	var brotliBuf bytes.Buffer
	brotliWriter := brotli.NewWriterLevel(&brotliBuf, brotli.BestCompression)
	_, err = brotliWriter.Write(identityData)
	if err != nil {
		return nil, fmt.Errorf("Failed to write brotli data: %v", err)
	}
	brotliWriter.Close()

	return &PreCompressedData{
		Gzip:     gzipBuf.Bytes(),
		Deflate:  deflateBuf.Bytes(),
		Brotli:   brotliBuf.Bytes(),
		Zstd:     zstdBuf.Bytes(),
		Identity: identityData,
	}, nil
}
