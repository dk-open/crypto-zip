package jetjson_test

import (
	"encoding/json"
	"fmt"
	"github.com/dk-open/crypto-zip/tools/jetjson"
	"github.com/dk-open/crypto-zip/types"
	gojson "github.com/goccy/go-json"
	jsoniterGo "github.com/json-iterator/go"
	"io"
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

var jsoniter = jsoniterGo.ConfigFastest

// InitializeTestReader reads the content of a text file and returns an io.Reader
func initializeTestReader(t *testing.T, filePath string) io.Reader {
	t.Helper() // Mark this as a helper function for cleaner test output

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	return file
}

// initializeBenchReader reads the content of a text file and returns an io.Reader
func initializeBenchReader(b *testing.B, filePath string) io.Reader {
	b.Helper() // Mark this as a helper function for cleaner test output

	file, err := os.Open(filePath)
	if err != nil {
		b.Fatalf("failed to open file: %v", err)
	}

	return file
}

type testStruct1 struct {
	Symbol   string              `json:"symbol"`
	BidPrice types.StringToFloat `json:"bidPrice"`
	AskPrice types.StringToFloat `json:"askPrice"`
}

func TestIterator(t *testing.T) {
	r := initializeTestReader(t, "test.json")
	si := jetjson.Decoder[testStruct1](r, 1)
	//si.Next()
	si.Read(func(item testStruct1) error {
		fmt.Println(item)
		return nil
	})
	fmt.Println(si)

}

func TestCPUProfile(t *testing.T) {
	// CPU profiling
	cpuProfile, err := os.Create("cpu.prof")
	if err != nil {
		t.Fatal("could not create CPU profile: ", err)
	}
	defer cpuProfile.Close()

	if err = pprof.StartCPUProfile(cpuProfile); err != nil {
		t.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	r := initializeTestReader(t, "test2.json")
	si := jetjson.Decoder[testStruct1](r, 1)
	//si.Next()
	si.Read(func(item testStruct1) error {
		return nil
	})
	fmt.Println(si)

	r = initializeTestReader(t, "test.json")
	si = jetjson.Decoder[testStruct1](r, 1)
	//si.Next()
	si.Read(func(item testStruct1) error {
		//fmt.Println(item)
		return nil
	})

	i := 0
	for i < 500 {
		r = initializeTestReader(t, "test.json")
		si = jetjson.Decoder[testStruct1](r, 1)
		//si.Next()
		si.Read(func(item testStruct1) error {
			//fmt.Println(item)
			return nil
		})
		i++
	}
	time.Sleep(3 * time.Second)

}

func TestGoJson1(t *testing.T) {
	path := "test.json"
	var res []testStruct
	r := initializeTestReader(t, path)
	if err := gojson.NewDecoder(r).Decode(&res); err != nil {
		t.Fatal(err)
	}
	for _, v := range res {
		fmt.Println(v.Symbol, v.BidPrice.Float(), v.AskPrice.Float())
	}
}

func TestGoJson(t *testing.T) {
	path := "test.json"
	//var res []testStruct
	r := initializeTestReader(t, path)

	d := gojson.NewDecoder(r)
	tk, err := d.Token()
	fmt.Println(tk, err)
	//tk, err = d.Token()
	//fmt.Println(tk, err)
	st := &testStruct1{}
	//var str string
	err = d.Decode(&st)
	fmt.Println(st, err)

	//if err := gojson.NewDecoder(r).Decode(&res); err != nil {
	//	t.Fatal(err)
	//}
	//for _, v := range res {
	//	fmt.Println(v.Symbol, v.BidPrice.Float(), v.AskPrice.Float())
	//}
}

type testStruct struct {
	Symbol   string              `json:"symbol"`
	BidPrice types.StringToFloat `json:"bidPrice"`
	AskPrice types.StringToFloat `json:"askPrice"`
}

func BenchmarkIterator(b *testing.B) {

	path := "test.json"
	b.Run("Decoder", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			r := initializeBenchReader(b, path)
			si := jetjson.Decoder[testStruct1](r, 1)
			si.Next()
			//_ = si
			si.Read(func(item testStruct1) error {
				return nil
			})
		}
		b.ReportAllocs()
	})

	var res []testStruct
	b.Run("Json", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := initializeBenchReader(b, path)
			if err := json.NewDecoder(r).Decode(&res); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("go-json", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := initializeBenchReader(b, path)
			if err := gojson.NewDecoder(r).Decode(&res); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportAllocs()
	})

	b.Run("json-iter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := initializeBenchReader(b, path)
			if err := jsoniter.NewDecoder(r).Decode(&res); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportAllocs()
	})
	//	github.com/goccy/go-json
}
