package compress

import (
	"bytes"
	"encoding/binary"
	"io"
)

type price struct {
	bid     uint64
	askDiff uint64
}

func Price(bid, askDiff uint64) IPrice {
	return &price{bid: bid, askDiff: askDiff}
}

func (p *price) Price(precision float64) (res [2]float64) {
	return [2]float64{float64(p.bid) / precision, float64(p.bid+p.askDiff) / precision}
}

type IPrice interface {
	Price(precision float64) [2]float64
	Write(buf *bytes.Buffer) error
	Bytes() []byte
}

// Write encodes the bid and askDiff into the provided bytes.Buffer
func (p *price) Write(buf *bytes.Buffer) error {
	if err := WriteVariant(buf, p.bid); err != nil {
		return err
	}
	return WriteVariant(buf, p.askDiff)
}

func PackPrice(buf io.ByteWriter, bid, askDiff uint64) error {
	if err := WriteVariant(buf, bid); err != nil {
		return err
	}
	if err := WriteVariant(buf, askDiff); err != nil {
		return err
	}
	return nil
}

func PriceZip(bid, askDiff uint64) []byte {
	var variant [2 * binary.MaxVarintLen64]byte
	n := binary.PutUvarint(variant[:], bid)
	n += binary.PutUvarint(variant[n:], askDiff)
	return variant[:n]
}

func (p *price) Bytes() []byte {
	var variant [2 * binary.MaxVarintLen64]byte

	n := binary.PutUvarint(variant[:], p.bid)
	n += binary.PutUvarint(variant[n:], p.askDiff)
	return variant[:n]
}

func WriteVariant(buf io.ByteWriter, x uint64) error {
	for x >= 0x80 {
		if err := buf.WriteByte(byte(x) | 0x80); err != nil {
			return err
		}
		x >>= 7
	}
	return buf.WriteByte(byte(x))
}
