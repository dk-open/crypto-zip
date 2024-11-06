package http

import (
	"context"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/dk-open/crypto-zip/tools/jetjson"
	"github.com/goccy/go-json"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"io"
	"log"
	"net/http"
)

type FetchFunc[TModel any] func(data *TModel) error
type IterateFetch[TModel any] func(f func(data TModel) error) error

type FetcherReader func(ctx context.Context, f func(ctx context.Context, reader io.Reader) error) error

func FetcherWithClient[TModel any](c *http.Client, method string, url string, headers ...HeaderOption) FetchFunc[TModel] {
	req, rErr := http.NewRequest(method, url, nil)
	if rErr != nil {
		log.Fatal(rErr)
	}

	for _, h := range headers {
		h(req)
	}

	return func(data *TModel) error {
		return fetchRequest(c, req, data)
	}
}

func Fetcher[TModel any](method string, url string) FetchFunc[TModel] {
	req, rErr := http.NewRequest(method, url, nil)
	if rErr != nil {
		log.Fatal(rErr)
	}
	req.Header.Set("Accept-Encoding", "br,gzip,deflate")

	return func(data *TModel) error {
		return fetchRequest(client, req, data)
	}
}

func FetchDefaultEncoded[TModel any](method string, url string, res *TModel) error {
	req, rErr := http.NewRequest(method, url, nil)
	if rErr != nil {
		log.Fatal(rErr)
	}
	req.Header.Set("Accept-Encoding", "br,gzip,deflate")

	return fetchRequest(http.DefaultClient, req, res)
}

func fetchRequest[TModel any](c *http.Client, req *http.Request, res *TModel) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid response. Error code %d", resp.StatusCode)
	}

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		rc, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer rc.Close()
		return json.NewDecoder(rc).Decode(res)
	case "deflate":
		rc, err := zlib.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer rc.Close()
		return json.NewDecoder(rc).Decode(res)
	case "br":
		return json.NewDecoder(brotli.NewReader(resp.Body)).Decode(res)
	default:
		return json.NewDecoder(resp.Body).Decode(res)
	}
}

func FetcherCustom(c *http.Client, method string, url string, headers ...HeaderOption) FetcherReader {
	req, rErr := http.NewRequest(method, url, nil)
	if rErr != nil {
		log.Fatal(rErr)
	}

	for _, h := range headers {
		h(req)
	}

	return func(ctx context.Context, f func(ctx context.Context, reader io.Reader) error) error {
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("Invalid response. Error code %d", resp.StatusCode)
		}

		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			rc, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer rc.Close()

			return f(ctx, rc)
		case "deflate":
			rc, err := zlib.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer rc.Close()
			return f(ctx, rc)
		case "br":
			return f(ctx, brotli.NewReader(resp.Body))
		default:
			return f(ctx, resp.Body)
		}
	}
}

func Iterator[TModel any](method string, url string, level int, headers ...HeaderOption) IterateFetch[TModel] {
	req, rErr := http.NewRequest(method, url, nil)
	if rErr != nil {
		log.Fatal(rErr)
	}

	for _, h := range headers {
		h(req)
	}

	return func(f func(data TModel) error) error {
		return fetchRequestIterator(client, req, level, f)
	}
}

func IteratorWithClient[TModel any](c *http.Client, method string, url string, level int, headers ...HeaderOption) IterateFetch[TModel] {
	req, rErr := http.NewRequest(method, url, nil)
	if rErr != nil {
		log.Fatal(rErr)
	}

	for _, h := range headers {
		h(req)
	}

	return func(f func(data TModel) error) error {
		return fetchRequestIterator(c, req, level, f)
	}
}

func fetchRequestIterator[TModel any](c *http.Client, req *http.Request, level int, f func(data TModel) error) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid response. Error code %d", resp.StatusCode)
	}

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		rc, rErr := gzip.NewReader(resp.Body)
		if rErr != nil {
			return rErr
		}
		defer rc.Close()
		return jetjson.Decoder[TModel](rc, level).Read(f)
	case "deflate":
		rc, rErr := zlib.NewReader(resp.Body)
		if rErr != nil {
			return rErr
		}
		defer rc.Close()
		return jetjson.Decoder[TModel](rc, level).Read(f)
	case "br":
		return jetjson.Decoder[TModel](brotli.NewReader(resp.Body), level).Read(f)
	default:
		return jetjson.Decoder[TModel](resp.Body, level).Read(f)
	}
}
