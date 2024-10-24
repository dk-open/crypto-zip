package jetjson

import (
	"io"
	"unsafe"
)

type Iterator struct {
	reader io.Reader
	buf    []byte
	err    error
	head   int
	tail   int

	//TODO: Improve this
	recordStart   int
	recordBuf     [128]byte
	recordBufSize int
	p             unsafe.Pointer
}

func NewIterator(reader io.Reader) *Iterator {
	res := &Iterator{reader: reader,
		buf: make([]byte, 1500),
	}
	res.p = res.bufptr()

	return res
}

func (b *Iterator) Error() error {
	return b.err
}

func (b *Iterator) Start() bool {
	for {
		c := char(b.p, b.head)
		switch c {
		case '{', '[':
			b.head += 1
			return true
		case nul:
			if !b.loadMore() {
				return false
			}
			continue
		}
		b.head += 1
	}
}

func (b *Iterator) Next() bool {
	for {
		switch char(b.p, b.head) {
		case ',', '}', ']':
			b.head += 1
			return true
		case nul:
			if !b.loadMore() {
				return false
			}
			continue
		}
		b.head += 1
	}
}

func (b *Iterator) End() bool {
	for {
		switch char(b.p, b.head) {
		case '}', ']':
			b.head += 1
			return true
		case nul:
			if !b.loadMore() {
				return false
			}
		}
		b.head += 1
		//fmt.Print(string(char(b.p, b.head)))
		continue
	}
}

// If use reflect.SliceHeader, data type is uintptr.
// In this case, Go compiler cannot trace reference created by newArray().
// So, define using unsafe.Pointer as data type
type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

const nul = '\000'

func (b *Iterator) bufptr() unsafe.Pointer {
	return (*sliceHeader)(unsafe.Pointer(&b.buf)).data
}

func char(ptr unsafe.Pointer, offset int) byte {
	return *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(offset)))
}

// Lookup table for whitespace characters
var whitespaceTable = [256]bool{
	' ':  true,
	'\n': true,
	'\t': true,
	'\r': true,
	':':  true,
}

func (b *Iterator) skipEmpty() byte {
LOOP:
	c := char(b.p, b.head)
	if whitespaceTable[c] {
		b.head++
		goto LOOP
	} else if c == nul {
		if !b.loadMore() {
			return nul
		}
		goto LOOP
	}
	return c
}

func (b *Iterator) takeValue() (res []byte) {
	if b.recordBufSize > 0 {
		//b.recordStart = 0
		copy(b.recordBuf[b.recordBufSize:], b.buf[:b.head])
		b.recordStart = 0
		return b.recordBuf[:b.recordBufSize+b.head]
		//return append(b.recordBuf[:b.recordBufSize], b.buf[:b.head]...)
	}
	res = b.buf[b.recordStart:b.head]
	b.recordStart = 0
	return res
}

func (b *Iterator) ReadKey() ([]byte, bool) {
	c := b.skipEmpty()
	b.recordStart = b.head
	b.recordBufSize = 0
	for {
		switch c {
		case ':':
			return b.takeValue(), true
		case nul:
			if !b.loadMore() {
				//return b.takeValue(), false
				return nil, false
			}
			//c = char(b.p, b.head)
		default:
			b.head++
		}
		c = char(b.p, b.head)
	}
}

// Lookup table for whitespace characters
var valueEndTable = [256]bool{
	'\n': true,
	'\t': true,
	'\r': true,
	' ':  true,
	',':  true,
	'}':  true,
	']':  true,
}

func (b *Iterator) ReadValue() ([]byte, bool) {
	c := b.skipEmpty()
	b.recordStart = b.head
	b.recordBufSize = 0
	for {
		if valueEndTable[c] {
			//fmt.Println(b.head, b.tail)
			return b.takeValue(), true
		} else if c == nul {
			if !b.loadMore() {
				return b.takeValue(), false
			}
		}

		b.head++
		c = char(b.p, b.head)
	}
}

func (b *Iterator) updateRecorder() {
	b.recordBufSize = b.tail - b.recordStart
	copy(b.recordBuf[:], b.buf[b.recordStart:b.tail])
}

//func (b *Iterator) loadNext() {
//
//}

//	func (b *Iterator) copyNext(){
//		buf:=<-b.nextBuf
//
// }
//
//	func (b *Iterator) readParallel() {
//		for {
//			select {
//
//			case <-b.readStart:
//				//fmt.Println("Do read")
//				nR, err := b.reader.Read(b.parallelBuf[:])
//				if err != nil {
//					b.err = err
//				}
//				//fmt.Println("Read done", nR)
//				b.readDone <- nR
//			}
//		}
//	}
//func (b *Iterator) readNext() int {
//	//fmt.Println("Read next")
//	n := <-b.readDone
//	//fmt.Println("Read next done", n)
//	copy(b.buf[:], b.parallelBuf[:n])
//
//	//go b.readParallel()
//	b.readStart <- true
//	//fmt.Println("Read next done2", n)
//	//fmt.Println(string(b.buf[:n]))
//	return n
//}
//
//func (b *Iterator) loadMore2() bool {
//	if b.recordStart > 0 && b.recordStart < b.tail {
//		b.updateRecorder()
//	}
//
//	if n := b.readNext(); n > 0 {
//		b.head = 0
//		b.tail = n
//		//os.Exit(1)
//		return true
//	}
//	return false
//
//}

func (b *Iterator) loadMore() bool {
	//fmt.Println("loadMore")
	if b.recordStart > 0 {
		if b.recordStart < b.tail {
			b.updateRecorder()
		}
		b.recordStart = 0
	}

	n, err := b.reader.Read(b.buf[:])
	//fmt.Println("load", n, err)
	if n == 0 {
		if err != nil {
			if b.err == nil {
				b.err = err
			}
		}
		return false
	} else {
		b.head = 0
		b.tail = n
		if n < 1500 {
			b.buf[n] = nul
		}
		return true
	}
}
