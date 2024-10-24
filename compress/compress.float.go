package compress

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

/*
	TODO: Refactor

-	Use variants.
-	Use data compression outsides
*/
func Encode(data []float64, precision uint64) ([]byte, error) {
	precisionF := float64(precision)
	firstValue := uint64(data[0] * precisionF)

	packedDeltas := make([]byte, 16)
	binary.BigEndian.PutUint64(packedDeltas[:8], precision)
	binary.BigEndian.PutUint64(packedDeltas[8:], firstValue)

	firstData := make([]byte, 8)
	binary.BigEndian.PutUint64(firstData, firstValue)

	prevValue := firstValue
	for i := 1; i < len(data); i++ {
		currentValue := uint64(data[i] * precisionF)
		delta := int64(currentValue) - int64(prevValue)

		p := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(p, delta)
		// Store delta in 2 bytes
		packedDeltas = append(packedDeltas, p[:n]...)

		prevValue = currentValue
	}

	var buf bytes.Buffer
	gz := zlib.NewWriter(&buf)

	_, err := gz.Write(packedDeltas)
	if err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(packedDeltas []byte) ([]float64, error) {
	reader, err := zlib.NewReader(bytes.NewReader(packedDeltas))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	packedDeltas, err = io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	precision := float64(binary.BigEndian.Uint64(packedDeltas[:8]))
	firstValue := binary.BigEndian.Uint64(packedDeltas[8:16])

	packedDeltas = packedDeltas[16:]
	var data = []float64{float64(firstValue) / precision}
	prevValue := int64(firstValue)

	i := 1
	for len(packedDeltas) > 0 {
		delta, n := binary.Varint(packedDeltas)
		prevValue = prevValue + delta
		data = append(data, float64(prevValue)/precision)
		packedDeltas = packedDeltas[n:]
		i++
	}
	return data, nil
}
