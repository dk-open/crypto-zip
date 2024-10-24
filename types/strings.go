package types

import (
	"bytes"
	"fmt"
	"github.com/valyala/fastjson/fastfloat"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type StringToFloat float64

func (foe StringToFloat) Float() float64 {
	return float64(foe)
}

var nullValue = []byte("null")
var emptyValue = []byte(`""`)

func (foe *StringToFloat) UnmarshalJSON(data []byte) error {
	// Handle null or empty data
	if len(data) == 0 || bytes.Equal(data, nullValue) || bytes.Equal(data, emptyValue) {
		*foe = 0.0
		return nil
	}

	if data[0] == '"' {
		*foe = StringToFloat(fastfloat.ParseBestEffort(BytesToString(data[1 : len(data)-1])))
		return nil
	}

	*foe = StringToFloat(fastfloat.ParseBestEffort(BytesToString(data)))
	return nil
}

type StringToUint64 uint64

func (uoe *StringToUint64) String() string {
	return fmt.Sprint(*uoe)
}

func (uoe *StringToUint64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Equal(data, nullValue) || bytes.Equal(data, emptyValue) {
		*uoe = 0
		return nil
	}

	if data[0] == '"' {
		data = data[1 : len(data)-1]
	}
	n, err := strconv.ParseUint(BytesToString(data), 10, 64)
	if err != nil {
		return err
	}
	*uoe = StringToUint64(n)
	return nil
}

type StringToTimeStampMs time.Time

func (t StringToTimeStampMs) String() string {
	return time.Time(t).Format(time.RFC3339)
}

func (t StringToTimeStampMs) Time() time.Time {
	return time.Time(t)
}

func (t StringToTimeStampMs) UnixMilli() int64 {
	return time.Time(t).UnixMilli()
}

func (uoe *StringToTimeStampMs) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Equal(data, nullValue) || bytes.Equal(data, emptyValue) {
		if uoe != nil {
			*uoe = StringToTimeStampMs(time.Time{})
		}
		return nil
	}

	num := strings.ReplaceAll(BytesToString(data), "\"", "")
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return err
	}
	*uoe = StringToTimeStampMs(time.UnixMilli(n))
	return nil
}

type TimestampToTime time.Time

func (tf TimestampToTime) String() string {
	return time.Time(tf).Format("2006-01-02T15:04:05.000Z")
}

func (tf *TimestampToTime) MarshalJSON() ([]byte, error) {
	str := time.Time(*tf).Format("2006-01-02T15:04:05.000Z")
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

func (tf *TimestampToTime) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		return nil
	}
	v, err := strconv.Atoi(BytesToString(data))
	if err != nil {
		return err
	}
	*tf = TimestampToTime(time.Unix(0, int64(v)*1000000))
	return nil
}

// BytesToString converts a byte slice to a string without allocation
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
