package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/hdt3213/rdb/model"
)

// rdbWith builds a tiny RDB file with one object of the given type code and key "k".
func rdbWith(typeCode byte, body ...[]byte) []byte {
	buf := []byte("REDIS0009")
	buf = append(buf, typeCode, 0x01, 'k')
	for _, b := range body {
		buf = append(buf, b...)
	}
	return append(buf, 0xff)
}

func length64(n uint64) []byte {
	b := make([]byte, 9)
	b[0] = len64Bit
	binary.BigEndian.PutUint64(b[1:], n)
	return b
}

func le32(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

func TestParseDeclaredLengthExceedsData(t *testing.T) {
	const huge = 1 << 40
	// intset: intSize 2, cardinality 1000, but no entries
	intset := append(le32(2), le32(1000)...)
	cases := []struct {
		name string
		data []byte
	}{
		{"string", rdbWith(0, length64(huge))},
		{"lzf", rdbWith(0, []byte{encodeLZFPrefix, 0x01}, length64(huge), []byte{0x00})},
		{"list", rdbWith(1, length64(huge))},
		{"set", rdbWith(2, length64(huge))},
		{"zset", rdbWith(3, length64(huge))},
		{"intset", rdbWith(11, []byte{byte(len(intset))}, intset)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			dec := NewDecoder(bytes.NewReader(c.data))
			err := dec.Parse(func(o model.RedisObject) bool { return true })
			if err == nil {
				t.Fatal("expect an error")
			}
		})
	}
}

func TestParseValkeySlotImportHugeRanges(t *testing.T) {
	data := []byte("VALKEY080")
	data = append(data, opCodeSlotImport, 0x01, 'j')
	data = append(data, length64(1<<40)...)
	data = append(data, 0xff)
	dec := NewDecoder(bytes.NewReader(data))
	if err := dec.Parse(func(o model.RedisObject) bool { return true }); err == nil {
		t.Fatal("expect an error")
	}
}

func TestCheckCount(t *testing.T) {
	cases := []struct {
		name  string
		count int64
		size  int
		ok    bool
	}{
		{"zero", 0, 0, true},
		{"fits", 5, 5, true},
		{"too many", 6, 5, false},
		{"negative", -1, 5, false},
		{"huge", 1 << 62, 5, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			err := checkCount(c.count, c.size)
			if (err == nil) != c.ok {
				t.Fatalf("checkCount(%d, %d) = %v", c.count, c.size, err)
			}
		})
	}
}

func TestCapHint(t *testing.T) {
	if capHint(5) != 5 || capHint(1024) != 1024 || capHint(1<<40) != 1024 {
		t.Fatal("unexpected capHint")
	}
}

func TestReadStreamEntryContentHugeCount(t *testing.T) {
	// listpack entries: count = 100000, deleted = 0 ; the buffer holds only a few bytes
	dec := NewDecoder(bytes.NewReader(nil))
	// entries: count (24-bit int 100000), deleted 0, master field number 0, end flag 0
	buf := []byte{0xf2, 0xa0, 0x86, 0x01, 0x05, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01}
	cursor := 0
	_, err := dec.readStreamEntryContent(buf, &cursor, &model.StreamId{})
	if err == nil || !strings.Contains(err.Error(), "does not fit") {
		t.Fatalf("expect a count error, got %v", err)
	}
}

func TestReadBytesAboveChunk(t *testing.T) {
	n := readChunk + 10
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i % 251)
	}
	t.Run("full data", func(t *testing.T) {
		dec := NewDecoder(bytes.NewReader(data))
		got, err := dec.readBytes(uint64(n))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, data) || cap(got) != n || dec.readCount != n {
			t.Fatalf("bad result: len %d cap %d count %d", len(got), cap(got), dec.readCount)
		}
	})
	truncated := []struct {
		name string
		size int
	}{
		{"partial data", readChunk + 5},
		{"ends at chunk boundary", readChunk},
	}
	for _, c := range truncated {
		t.Run(c.name, func(t *testing.T) {
			dec := NewDecoder(bytes.NewReader(data[:c.size]))
			if _, err := dec.readBytes(uint64(n)); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("expect io.ErrUnexpectedEOF, got %v", err)
			}
		})
	}
}
