package core

import (
	"bytes"
	"github.com/hdt3213/rdb/model"
	"math"
	"strconv"
	"testing"
)

func TestListEncoding(t *testing.T) {
	var list [][]byte
	for i := 0; i < 1021; i++ { // pick a prime number
		list = append(list, []byte(RandString(128)))
	}
	listMap := map[string][][]byte{
		"a": {
			[]byte("a"), []byte("b"), []byte("c"), []byte("d"),
		},
		"1": {
			[]byte("1"), []byte("2"), []byte("3"), []byte("4"),
		},
		"001": {
			[]byte("001"), []byte("0x11"), []byte("000"), []byte("0"),
		},
		"0": {
			[]byte("0x11"), []byte("001"), []byte("11111111"), []byte("1"),
		},
		"large": list,
	}
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	err := enc.WriteHeader()
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.WriteDBHeader(0, uint64(len(listMap)), 0)
	if err != nil {
		t.Error(err)
		return
	}
	for k, v := range listMap {
		err = enc.WriteListObject(k, v)
		if err != nil {
			t.Error(err)
			return
		}
	}
	err = enc.WriteEnd()
	if err != nil {
		t.Error(err)
		return
	}
	dec := NewDecoder(buf).WithSpecialOpCode()
	err = dec.Parse(func(object model.RedisObject) bool {
		switch o := object.(type) {
		case *model.ListObject:
			expect := listMap[o.GetKey()]
			if len(expect) != o.GetElemCount() {
				t.Errorf("list %s has wrong element count", o.GetKey())
				return true
			}
			for i, expectV := range expect {
				actualV := o.Values[i]
				if !bytes.Equal(expectV, actualV) {
					t.Errorf("list %s has element at index %d", o.GetKey(), i)
					return true
				}
			}
		}
		return true
	})
	if err != nil {
		t.Error(err)
	}
}

func TestZipListEncoding(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf).SetListZipListOpt(64, 64)
	list := []string{
		"",
		"0",
		"1",
		"13",
		"127",
		"32766",
		"8388607",
		"16777216",
		"2147483647",
		"21474836471",
		"a",
		"abc",
		"007",
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
		RandString(60),
		RandString(1638),
		RandString(10000),
	}
	err := enc.writeZipList(list)
	if err != nil {
		t.Error(err)
		return
	}
	dec := NewDecoder(buf)
	actual, err := dec.readZipList()
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != len(actual) {
		t.Error("wrong result size")
		return
	}
	for i, expectV := range list {
		actualV := string(actual[i])
		if expectV != actualV {
			t.Errorf("illegal value at %d", i)
		}
	}
}
