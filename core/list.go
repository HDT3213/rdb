package core

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

const (
	zipStr06B = 0
	zipStr14B = 1
	zipStr32B = 2

	zipInt04B = 0x0f
	zipInt08B = 0xfe        // 11111110
	zipInt16B = 0xc0 | 0<<4 // 11000000
	zipInt24B = 0xc0 | 3<<4 // 11110000
	zipInt32B = 0xc0 | 1<<4 // 11010000
	zipInt64B = 0xc0 | 2<<4 //11100000

	zipBigPrevLen = 0xfe
)

func (dec *Decoder) readList() ([][]byte, error) {
	size64, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	size := int(size64)
	values := make([][]byte, 0, size)
	for i := 0; i < size; i++ {
		val, err := dec.readString()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

func (dec *Decoder) readZipList() ([][]byte, error) {
	buf, err := dec.readString()
	if err != nil {
		return nil, err
	}
	cursor := 0
	size := readZipListLength(buf, &cursor)
	entries := make([][]byte, 0, size)
	for i := 0; i < size; i++ {
		entry, err := dec.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (dec *Decoder) readZipListEntry(buf []byte, cursor *int) (result []byte, err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic: %v", err)
		}
	}()
	prevLen := buf[*cursor]
	*cursor++
	if prevLen == zipBigPrevLen {
		*cursor += 4
	}
	header := buf[*cursor]
	*cursor++
	typ := header >> 6
	switch typ {
	case zipStr06B:
		length := int(header & 0x3f)
		result, err = readBytes(buf, cursor, length)
		return
	case zipStr14B:
		b := buf[*cursor]
		*cursor++
		length := (int(header&0x3f) << 8) | int(b)
		result, err = readBytes(buf, cursor, length)
		return
	case zipStr32B:
		var lenBytes []byte
		lenBytes, err = readBytes(buf, cursor, 4)
		if err != nil {
			return
		}
		length := int(binary.BigEndian.Uint32(lenBytes))
		result, err = readBytes(buf, cursor, length)
		return
	}
	switch header {
	case zipInt08B:
		var b byte
		b, err = readByte(buf, cursor)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(int8(b)), 10))
		return
	case zipInt16B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 2)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(int16(binary.LittleEndian.Uint16(bs))), 10))
		return
	case zipInt32B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 4)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))), 10))
		return
	case zipInt64B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 8)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(binary.LittleEndian.Uint64(bs)), 10))
		return
	case zipInt24B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 3)
		if err != nil {
			return
		}
		bs = append([]byte{0}, bs...)
		result = []byte(strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))>>8), 10))
		return
	}
	if header>>4 == zipInt04B {
		result = []byte(strconv.FormatInt(int64(header&0x0f)-1, 10))
		return
	}
	return nil, fmt.Errorf("unknown entry header")
}

func (dec *Decoder) readQuickList() ([][]byte, error) {
	size, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	entries := make([][]byte, 0)
	for i := 0; i < int(size); i++ {
		page, err := dec.readZipList()
		if err != nil {
			return nil, err
		}
		entries = append(entries, page...)
	}
	return entries, nil
}
