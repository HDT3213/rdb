package core

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestLengthEncoding(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	lens := []uint64{1 << 5, 1 << 13, 1 << 31, 1 << 63}
	for _, v := range lens {
		err := enc.writeLength(v)
		if err != nil {
			t.Error(err)
			return
		}
	}
	dec := NewDecoder(buf)
	for _, expect := range lens {
		actual, special, err := dec.readLength()
		if err != nil {
			t.Error(err)
			return
		}
		if special {
			t.Error("expect normal actual special")
			return
		}
		if actual != expect {
			t.Errorf("expect %d, actual %d", expect, actual)
		}
	}
}

func TestStringEncoding(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	strList := []string{
		"",
		"abc",
		"12",
		"32766",
		"007",
		"0",
		"+0",
		"-0",
		"+1",
		"-1",
		"0x11",
		"0o00",
		strconv.Itoa(math.MaxInt8),
		strconv.Itoa(math.MinInt8),
		strconv.Itoa(math.MaxInt16),
		strconv.Itoa(math.MinInt16),
		strconv.Itoa(math.MaxInt32),
		strconv.Itoa(math.MinInt32),
		strconv.Itoa(math.MinInt32) + "1",
		strconv.Itoa(math.MaxInt64),
		strconv.Itoa(math.MinInt64),
		strconv.Itoa(math.MinInt64) + "1",
		RandString(20000),
	}
	for _, str := range strList {
		err := enc.writeString(str)
		if err != nil {
			t.Error(err)
			return
		}
	}
	dec := NewDecoder(buf)
	for _, expect := range strList {
		actual, err := dec.readString()
		if err != nil {
			t.Error(err)
			continue
		}
		if string(actual) != expect {
			t.Errorf("expect %s, actual %s", expect, string(actual))
		}
	}
}

func TestLZFString(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf).EnableCompress()
	var strList []string
	for i := 0; i < 10; i++ {
		strList = append(strList, strings.Repeat(RandString(128), 10))
	}
	for _, str := range strList {
		err := enc.writeString(str)
		if err != nil {
			t.Error(err)
			return
		}
	}
	dec := NewDecoder(buf)
	for _, expect := range strList {
		actual, err := dec.readString()
		if err != nil {
			t.Error(err)
			continue
		}
		if string(actual) != expect {
			t.Errorf("expect %s, actual %s", expect, string(actual))
		}
	}
}
