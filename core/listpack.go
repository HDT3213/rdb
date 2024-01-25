package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
)

// readListPack returns: list of entry, list of entry size, error
func (dec *Decoder) readListPack() ([][]byte, []uint32, error) {
	buf, err := dec.readString()
	if err != nil {
		return nil, nil, err
	}
	cursor := 0
	size := readListPackLength(buf, &cursor)
	entries := make([][]byte, 0, size)
	entrySizes := make([]uint32, 0, size)
	for i := 0; i < size; i++ {
		str, intval, length, err := dec.readListPackEntry(buf, &cursor)
		if err != nil {
			return nil, nil, err
		}
		if str == nil {
			str = []byte(strconv.FormatInt(intval, 10))
		}
		entries = append(entries, str)
		entrySizes = append(entrySizes, length)
	}
	return entries, entrySizes, nil
}

func readListPackLength(buf []byte, cursor *int) int {
	start := *cursor + 4
	end := start + 2
	// list pack buf: [0, 4] -> total bytes, [4:6] -> entry count
	size := int(binary.LittleEndian.Uint16(buf[start:end]))
	*cursor += 6
	return size
}

func getBackLen(elementLen uint32) uint32 {
	if elementLen <= 127 {
		return 1
	} else if elementLen < (1<<14)-1 {
		return 2
	} else if elementLen < (1<<21)-1 {
		return 3
	} else if elementLen < (1<<28)-1 {
		return 4
	} else {
		return 5
	}
}


// readListPackEntry returns: string content, int content, entry length(encoding+content+backlen), error
func (dec *Decoder) readListPackEntry(buf []byte, cursor *int) ([]byte, int64, uint32, error) {
	header, err := readByte(buf, cursor)
	if err != nil {
		return nil, 0, 0, err
	}
	switch header >> 6 {
	case 0, 1: // 0xxxxxxx, uint7
		result := int64(int8(header))
		var contentLen uint32 = 1
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return nil, result, contentLen + backlen, nil
	case 2: // 10xxxxxx + content, string(len<=63)
		strLen := int(header & 0x3f)
		result, err := readBytes(buf, cursor, strLen)
		if err != nil {
			return nil, 0, 0, err
		}
		var contentLen = uint32(1 + strLen)
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return result, 0, contentLen + backlen, nil
	}
	// assert header == 11xxxxxx
	switch header >> 4 {
	case 12, 13: // 110xxxxx yyyyyyyy, int13
		// see https://github.com/CN-annotation-team/redis7.0-chinese-annotated/blob/fba43c524524cbdb54955a28af228b513420d78d/src/listpack.c#L586
		next, err := readByte(buf, cursor)
		if err != nil {
			return nil, 0, 0, err
		}
		val := ((uint(header) & 0x1F) << 8) | uint(next)
		if val >= uint(1<<12) {
			val = -(8191 - val) - 1 // val is uint, must use -(8191 - val), val - 8191 will cause overflow
		}
		result := int64(val)
		var contentLen uint32 = 2
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return nil, result, contentLen + backlen, nil
	case 14: // 1110xxxx yyyyyyyy + content, string(len < 1<<12)
		dec.buffer[0] = header & 0x0f
		dec.buffer[1], err = readByte(buf, cursor)
		if err != nil {
			return nil, 0, 0, err
		}
		strLen := binary.BigEndian.Uint16(dec.buffer[:2])
		result, err := readBytes(buf, cursor, int(strLen))
		if err != nil {
			return nil, 0, 0, err
		}
		var contentLen = uint32(2 + strLen)
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return result, 0, contentLen + backlen, nil
	}
	// assert header == 1111xxxx
	switch header & 0x0f {
	case 0: // 11110000 aaaaaaaa bbbbbbbb cccccccc dddddddd + content, string(len < 1<<32)
		var lenBytes []byte
		lenBytes, err = readBytes(buf, cursor, 4)
		if err != nil {
			return nil, 0, 0, err
		}
		strLen := int(binary.LittleEndian.Uint32(lenBytes))
		result, err := readBytes(buf, cursor, strLen)
		if err != nil {
			return nil, 0, 0, err
		}
		var contentLen = uint32(1 + 4 + strLen)
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return result, 0, contentLen + backlen, nil
	case 1: // 11110001 aaaaaaaa bbbbbbbb, int16
		var bs []byte
		bs, err = readBytes(buf, cursor, 2)
		if err != nil {
			return nil, 0, 0, err
		}
		result := int64(int16(binary.LittleEndian.Uint16(bs)))
		var contentLen uint32 = 3
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return nil, result, contentLen + backlen, nil
	case 2: // 11110010 aaaaaaaa bbbbbbbb cccccccc, int24
		var bs []byte
		bs, err = readBytes(buf, cursor, 3)
		if err != nil {
			return nil, 0, 0, err
		}
		bs = append([]byte{0}, bs...)
		result := int64(int32(binary.LittleEndian.Uint32(bs))>>8)
		var contentLen uint32 = 4
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return nil, result, contentLen + backlen, nil
	case 3: // 1111 0011 -> int32
		var bs []byte
		bs, err = readBytes(buf, cursor, 4)
		if err != nil {
			return nil, 0, 0, err
		}
		result := int64(int32(binary.LittleEndian.Uint32(bs)))
		var contentLen uint32 = 5
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return nil, result, contentLen + backlen, nil
	case 4: // 11110100 8Byte -> int64
		var bs []byte
		bs, err = readBytes(buf, cursor, 8)
		if err != nil {
			return nil, 0, 0, err
		}
		result := int64(binary.LittleEndian.Uint64(bs))
		var contentLen uint32 = 9
		backlen := getBackLen(contentLen)
		*cursor += int(backlen)
		return nil, result, contentLen + backlen, nil
	case 15: // 11111111 -> end
		return nil, 0, 0, errors.New("unexpected end")
	}
	return nil, 0, 0, fmt.Errorf("unknown entry header")
}

// readListPackEntryAsString return a string representation of entry
// It means if the entry is a integer, then format it as string
func (dec *Decoder) readListPackEntryAsString(buf []byte, cursor *int) ([]byte, error) {
	str, intval, _, err := dec.readListPackEntry(buf, cursor)
	if err != nil {
		return nil, fmt.Errorf("read from failed: %v", err)
	}
	if str != nil {
		return str, nil
	}
	str = []byte(strconv.FormatInt(intval, 10))
	return str, nil
}

func (dec *Decoder) readListPackEntryAsInt(buf []byte, cursor *int) (int64, error) {
	str, intval, _, err := dec.readListPackEntry(buf, cursor)
	if err != nil {
		return 0, fmt.Errorf("read from failed: %v", err)
	}
	if str != nil {
		return 0, fmt.Errorf("%s is not a integer", string(str))
	}
	return intval, nil
}
