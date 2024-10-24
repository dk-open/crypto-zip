package types_test

import (
	"github.com/dk-open/crypto-zip/types"
	"testing"
)

func BenchmarkUnmarshalJSONOriginal(b *testing.B) {
	testData := []string{
		`"122223.456"`,
		`12333.456`,
		`"0"`,
		`0`,
		`"null"`,
		`null`,
		`""`,
		`"-789.1011"`,
	}

	var foe types.StringToFloat

	for _, data := range testData {
		b.Run("Data="+data, func(b *testing.B) {
			bytesData := []byte(data)
			for i := 0; i < b.N; i++ {
				if err := foe.UnmarshalJSON(bytesData); err != nil {
					b.Error(err)
				}
			}
			b.ReportAllocs()
		})

	}
}
