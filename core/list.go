package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

const (
	zipStr06B = 0
	zipStr14B = 1
	zipStr32B = 2

	zipInt04B = 0x0f        // high 4 bits of Int 04 encoding
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

func (enc *Encoder) WriteListObject(key string, values [][]byte, options ...interface{}) error {
	err := enc.beforeWriteObject(options...)
	if err != nil {
		return err
	}
	ok, err := enc.tryWriteListZipList(key, values, options...)
	if err != nil {
		return err
	}
	if !ok {
		err = enc.writeQuickList(key, values, options...)
		if err != nil {
			return err
		}
	}
	enc.state = writtenObjectState
	return nil
}

func (enc *Encoder) tryWriteListZipList(key string, values [][]byte, options ...interface{}) (bool, error) {
	if len(values) > enc.listZipListOpt.getMaxEntries() {
		return false, nil
	}
	strList := make([]string, 0, len(values))
	maxValue := enc.listZipListOpt.getMaxValue()
	for _, v := range values {
		if len(v) > maxValue {
			return false, nil
		}
		strList = append(strList, string(v))
	}
	err := enc.write([]byte{typeListZipList})
	if err != nil {
		return true, err
	}
	err = enc.writeString(key)
	if err != nil {
		return true, err
	}
	err = enc.writeZipList(strList)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (enc *Encoder) writeQuickList(key string, values [][]byte, options ...interface{}) error {
	var pages [][]string
	pageSize := 0
	var curPage []string
	for _, value := range values {
		curPage = append(curPage, string(value))
		pageSize += len(value)
		if pageSize >= enc.listZipListSize {
			pageSize = 0
			pages = append(pages, curPage)
			curPage = nil
		}
	}
	if len(curPage) > 0 {
		pages = append(pages, curPage)
	}
	err := enc.write([]byte{typeListQuickList})
	if err != nil {
		return err
	}
	err = enc.writeString(key)
	if err != nil {
		return err
	}
	err = enc.writeLength(uint64(len(pages)))
	if err != nil {
		return err
	}
	for _, page := range pages {
		err = enc.writeZipList(page)
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeZipListEntry(prevLen uint32, val string) []byte {
	buf := bytes.NewBuffer(nil)
	// encode prevLen
	if prevLen < zipBigPrevLen {
		buf.Write([]byte{byte(prevLen)})
	} else {
		buf.Write([]byte{0xfe})
		buf0 := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf0, prevLen)
		buf.Write(buf0)
	}
	// try int encoding
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		// use int encoding
		if intVal >= 0 && intVal <= 12 {
			buf.Write([]byte{0xf0 | byte(intVal+1)})
		} else if intVal >= math.MinInt8 && intVal <= math.MaxInt8 {
			// bytes.Buffer never failed
			buf.Write([]byte{byte(zipInt08B), byte(intVal)})
		} else if intVal >= minInt24 && intVal <= maxInt24 {
			buffer := make([]byte, 4)
			binary.LittleEndian.PutUint32(buffer, uint32(intVal))
			buf.Write([]byte{byte(zipInt24B)})
			buf.Write(buffer[0:3])
		} else if intVal >= math.MinInt32 && intVal <= math.MaxInt32 {
			buffer := make([]byte, 4)
			binary.LittleEndian.PutUint32(buffer, uint32(intVal))
			buf.Write([]byte{byte(zipInt32B)})
			buf.Write(buffer)
		} else {
			buffer := make([]byte, 8)
			binary.LittleEndian.PutUint64(buffer, uint64(intVal))
			buf.Write([]byte{byte(zipInt64B)})
			buf.Write(buffer)
		}
		return buf.Bytes()
	}
	// use string encoding
	if len(val) <= maxUint6 {
		buf.Write([]byte{byte(len(val))}) // 00 + xxxxxx
	} else if len(val) <= maxUint14 {
		buf.Write([]byte{byte(len(val)>>8) | len14BitMask, byte(len(val))})
	} else if len(val) <= math.MaxUint32 {
		buffer := make([]byte, 8)
		binary.LittleEndian.PutUint32(buffer, uint32(len(val)))
		buf.Write([]byte{0x80})
		buf.Write(buffer)
	} else {
		panic("too large string")
	}
	buf.Write([]byte(val))
	return buf.Bytes()
}

func (enc *Encoder) writeZipList(values []string) error {
	buf := make([]byte, 10) // reserve 10 bytes for zip list header
	zlBytes := 11           // header(10bytes) + zl end(1byte)
	zlTail := 10
	var prevLen uint32
	for i, value := range values {
		entry := encodeZipListEntry(prevLen, value)
		buf = append(buf, entry...)
		prevLen = uint32(len(entry))
		zlBytes += len(entry)
		if i < len(values)-1 {
			zlTail += len(entry)
		}
	}
	buf = append(buf, 0xff)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(zlBytes))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(zlTail))
	binary.LittleEndian.PutUint16(buf[8:10], uint16(len(values)))
	return enc.writeNanString(string(buf))
}
