package parser

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

const (
	ZipStr06B = 0
	ZipStr14B = 1
	ZipStr32B = 2

	ZipInt04B = 0x0f
	ZipInt08B = 0xfe        // 11111110
	ZipInt16B = 0xc0 | 0<<4 // 11000000
	ZipInt24B = 0xc0 | 3<<4 // 11110000
	ZipInt32B = 0xc0 | 1<<4 // 11010000
	ZipInt64B = 0xc0 | 2<<4 //11100000

	ZipBigPrevLen = 0xfe
)

func (p *Parser) readList() ([][]byte, error) {
	size64, _, err := p.readLength()
	if err != nil {
		return nil, err
	}
	size := int(size64)
	values := make([][]byte, 0, size)
	for i := 0; i < size; i++ {
		val, err := p.readString()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

func (p *Parser) readZipList() ([][]byte, error) {
	buf, err := p.readString()
	if err != nil {
		return nil, err
	}
	cursor := 0
	size := readZipListLength(buf, &cursor)
	entries := make([][]byte, 0, size)
	for i := 0; i < size; i++ {
		entry, err := p.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (p *Parser) readZipListEntry(buf []byte, cursor *int) (result []byte, err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic: %v", err)
		}
	}()
	prevLen := buf[*cursor]
	*cursor++
	if prevLen == ZipBigPrevLen {
		*cursor += 4
	}
	header := buf[*cursor]
	*cursor++
	typ := header >> 6
	switch typ {
	case ZipStr06B:
		length := int(header & 0x3f)
		result, err = readBytes(buf, cursor, length)
		return
	case ZipStr14B:
		b := buf[*cursor]
		*cursor++
		length := (int(header&0x3f) << 8) | int(b)
		result, err = readBytes(buf, cursor, length)
		return
	case ZipStr32B:
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
	case ZipInt08B:
		var b byte
		b, err = readByte(buf, cursor)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(int8(b)), 10))
		return
	case ZipInt16B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 2)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(int16(binary.LittleEndian.Uint16(bs))), 10))
		return
	case ZipInt32B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 4)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))), 10))
		return
	case ZipInt64B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 8)
		if err != nil {
			return
		}
		result = []byte(strconv.FormatInt(int64(binary.LittleEndian.Uint64(bs)), 10))
		return
	case ZipInt24B:
		var bs []byte
		bs, err = readBytes(buf, cursor, 3)
		if err != nil {
			return
		}
		bs = append([]byte{0}, bs...)
		result = []byte(strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))>>8), 10))
		return
	}
	if header>>4 == ZipInt04B {
		result = []byte(strconv.FormatInt(int64(header&0x0f)-1, 10))
		return
	}
	return nil, fmt.Errorf("unknown entry header")
}

func (p *Parser) readQuickList() ([][]byte, error) {
	size, _, err := p.readLength()
	if err != nil {
		return nil, err
	}
	entries := make([][]byte, 0)
	for i := 0; i < int(size); i++ {
		page, err := p.readZipList()
		if err != nil {
			return nil, err
		}
		entries = append(entries, page...)
	}
	return entries, nil
}
