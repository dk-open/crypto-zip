package jetjson

import (
	"github.com/dk-open/crypto-zip/types"
	"github.com/valyala/fastjson/fastfloat"
	"io"
	"reflect"
	"unsafe"
)

type IDecoder[T any] interface {
	Next() bool
	Read(f func(item T) error) error
}

type decoder[T any] struct {
	iter *Iterator
}

func Decoder[T any](buf io.Reader, level int) IDecoder[T] {
	res := &decoder[T]{iter: NewIterator(buf)}
	for i := 0; i < level; i++ {
		res.Next()
	}
	return res
}

func stringUpdater(addr unsafe.Pointer) func(data []byte) {
	var stringVal [64]byte

	return func(data []byte) {
		data = data[1 : len(data)-1]
		copy(stringVal[:], data[:])
		*(*string)(addr) = string(stringVal[:len(data)])
	}
}

func float64Updater(addr unsafe.Pointer) func(data []byte) {
	return func(data []byte) {
		if data[0] == '"' {
			*(*float64)(addr) = fastfloat.ParseBestEffort(types.BytesToString(data[1 : len(data)-1]))
			return
		}
		*(*float64)(addr) = fastfloat.ParseBestEffort(types.BytesToString(data))
	}
}

func (s *decoder[T]) Next() bool {
	return s.iter.Start()
}

func (s *decoder[T]) Read(callBack func(item T) error) error {
	fieldUpdater := map[string]func([]byte){}

	var item T
	vv := reflect.ValueOf(&item).Elem()
	//reflect.ValueOf()
	for i := 0; i < vv.NumField(); i++ {
		fl := vv.Type().Field(i)
		fName := fl.Name
		if tag := fl.Tag.Get("json"); tag != "" {
			fName = tag
		}

		switch fl.Type.Kind() {
		case reflect.String:
			//fields = append(fields, []byte(fName))
			fieldUpdater[fName] = stringUpdater(unsafe.Pointer(vv.Field(i).UnsafeAddr()))
		case reflect.Float64:
			fieldUpdater[fName] = float64Updater(unsafe.Pointer(vv.Field(i).UnsafeAddr()))
		default:
			panic("unhandled default case")
		}
	}
	num := len(fieldUpdater)
	var matched int
	for s.iter.Start() {
		matched = num
		for matched > 0 {
			if key, kok := s.iter.ReadKey(); kok {

				//fmt.Println("key", string(key), s.iter.head, s.iter.tail)
				if fu, ok := fieldUpdater[types.BytesToString(key[1:len(key)-1])]; ok {
					if val, vok := s.iter.ReadValue(); vok {
						fu(val)
						matched--
					}
				}
			}

			if noNext := s.iter.Next(); !noNext {
				break
			}
		}
		if err := callBack(item); err != nil {
			return err
		}
		s.iter.End()
	}
	return nil
}
