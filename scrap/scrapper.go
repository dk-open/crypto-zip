package scrap

import (
	"bytes"
	"context"
)

type IScrapper interface {
	Scrap(ctx context.Context, buf *bytes.Buffer) error
}

type IPriceWriter interface {
	Write(name string, bid, ask float64) error
}
