package parser

import (
	"encoding/binary"
	"errors"
)

/*
	if hlen <= ZIPMAP_VALUE_MAX_FREE
*/

func (p *Parser) readHashMap() (map[string][]byte, error) {
	size, _, err := p.readLength()
	if err != nil {
		return nil, err
	}
	m := make(map[string][]byte)
	for i := 0; i < int(size); i++ {
		field, err := p.readString()
		if err != nil {
			return nil, err
		}
		value, err := p.readString()
		if err != nil {
			return nil, err
		}
		m[string(field)] = value
	}
	return m, nil
}

func (p *Parser) readZipMapHash() (map[string][]byte, error) {
	buf, err := p.readString()
	if err != nil {
		return nil, err
	}
	cursor := 0
	bLen, err := readByte(buf, &cursor)
	if err != nil {
		return nil, err
	}
	length := int(bLen)
	if bLen > 254 {
		//todo: scan once
		length, err = countZipMapEntries(buf, &cursor)
		if err != nil {
			return nil, err
		}
		length /= 2
	}
	m := make(map[string][]byte)
	for i := 0; i < length; i++ {
		fieldB, err := readZipMapEntry(buf, &cursor, false)
		if err != nil {
			return nil, err
		}
		field := string(fieldB)
		value, err := readZipMapEntry(buf, &cursor, true)
		if err != nil {
			return nil, err
		}
		m[field] = value
	}
	return m, nil
}

// return: len, free, error
func readZipMapEntryLen(buf []byte, cursor *int, readFree bool) (int, int, error) {
	b, err := readByte(buf, cursor)
	if err != nil {
		return 0, 0, err
	}
	switch b {
	case 253:
		bs, err := readBytes(buf, cursor, 5)
		if err != nil {
			return 0, 0, err
		}
		length := int(binary.BigEndian.Uint32(bs))
		free := int(bs[4])
		return length, free, nil
	case 254:
		return 0, 0, errors.New("illegal zip map item length")
	case 255:
		return -1, 0, nil
	default:
		var free byte
		if readFree {
			free, err = readByte(buf, cursor)
		}
		return int(b), int(free), err
	}
}

func readZipMapEntry(buf []byte, cursor *int, readFree bool) ([]byte, error) {
	length, free, err := readZipMapEntryLen(buf, cursor, readFree)
	if err != nil {
		return nil, err
	}
	if length == -1 {
		return nil, nil
	}
	value, err := readBytes(buf, cursor, length)
	if err != nil {
		return nil, err
	}
	*cursor += free // skip free bytes
	return value, nil
}

func countZipMapEntries(buf []byte, cursor *int) (int, error) {
	n := 0
	for {
		readFree := n%2 != 0
		length, free, err := readZipMapEntryLen(buf, cursor, readFree)
		if err != nil {
			return 0, err
		}
		if length == -1 {
			break
		}
		*cursor += length + free
		n++
	}
	*cursor = 0 // reset cursor
	return n, nil
}

func (p *Parser) readZipListHash() (map[string][]byte, error) {
	buf, err := p.readString()
	if err != nil {
		return nil, err
	}
	cursor := 0
	size := readZipListLength(buf, &cursor)
	m := make(map[string][]byte)
	for i := 0; i < size; i += 2 {
		key, err := p.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		val, err := p.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		m[string(key)] = val
	}
	return m, nil
}
