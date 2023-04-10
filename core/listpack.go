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
		entry, length, err := dec.readListPackEntry(buf, &cursor)
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, entry)
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

func readVarInt(buf []byte, cursor *int) uint32 {
	var v uint32
	shift := 0
	for *cursor < len(buf) {
		x := buf[*cursor]
		*cursor++
		v |= uint32(x&0x7f) << shift
		shift += 7
		if x&0x80 == 0 {
			break
		}
	}
	return v
}

// readListPackEntry returns: content(string), length, error
func (dec *Decoder) readListPackEntry(buf []byte, cursor *int) ([]byte, uint32, error) {
	header, err := readByte(buf, cursor)
	if err != nil {
		return nil, 0, err
	}
	var result []byte
	var length uint32
	switch header >> 6 {
	case 0, 1: // 0xxx xxxx -> uint7 [0, 127]
		result = []byte(strconv.FormatInt(int64(int8(header)), 10))
		length = readVarInt(buf, cursor) // read element length
		return result, length, nil
	case 2: // 10xx xxxx -> str, len<= 63
		strLen := int(header & 0x3f)
		result, err := readBytes(buf, cursor, strLen)
		if err != nil {
			return nil, 0, err
		}
		length = readVarInt(buf, cursor) // read element length
		return result, length, nil
	}
	// assert header == 11xx xxxx
	switch header >> 4 {
	case 12, 13: // 110x xxxx -> int13
		// see https://github.com/CN-annotation-team/redis7.0-chinese-annotated/blob/fba43c524524cbdb54955a28af228b513420d78d/src/listpack.c#L586
		next, err := readByte(buf, cursor)
		if err != nil {
			return nil, 0, err
		}
		val := ((uint(header) & 0x1F) << 8) | uint(next)
		if val >= uint(1<<12) {
			val = -(8191 - val) - 1 // val is uint, must use -(8191 - val), val - 8191 will cause overflow
		}
		result = []byte(strconv.FormatInt(int64(val), 10))
		length = readVarInt(buf, cursor) // read element length
		return result, length, nil
	case 14: // 1110 xxxx -> str, type(len) == uint12
		dec.buffer[0] = header & 0x0f
		dec.buffer[1], err = readByte(buf, cursor)
		strLen := binary.BigEndian.Uint16(dec.buffer[:2])
		result, err := readBytes(buf, cursor, int(strLen))
		if err != nil {
			return nil, 0, err
		}
		length = readVarInt(buf, cursor) // read element length
		return result, length, nil
	}
	// assert header == 1111 xxxx
	switch header & 0x0f {
	case 0: // 1111 0000 -> str, 4 bytes len
		var lenBytes []byte
		lenBytes, err = readBytes(buf, cursor, 4)
		if err != nil {
			return nil, 0, err
		}
		strLen := int(binary.BigEndian.Uint32(lenBytes))
		result, err := readBytes(buf, cursor, strLen)
		if err != nil {
			return nil, 0, err
		}
		length = readVarInt(buf, cursor) // read element length
		return result, length, nil
	case 1: // 1111 0001 -> int16
		var bs []byte
		bs, err = readBytes(buf, cursor, 2)
		if err != nil {
			return nil, 0, err
		}
		result = []byte(strconv.FormatInt(int64(int16(binary.LittleEndian.Uint16(bs))), 10))
		length = readVarInt(buf, cursor)
		return result, length, nil
	case 2: // 1111 0010 -> int24
		var bs []byte
		bs, err = readBytes(buf, cursor, 3)
		if err != nil {
			return nil, 0, err
		}
		bs = append([]byte{0}, bs...)
		result = []byte(strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))>>8), 10))
		length = readVarInt(buf, cursor)
		return result, length, nil
	case 3: // 1111 0011 -> int32
		var bs []byte
		bs, err = readBytes(buf, cursor, 4)
		if err != nil {
			return nil, 0, err
		}
		result = []byte(strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(bs))), 10))
		length = readVarInt(buf, cursor)
		return result, length, nil
	case 4: // 1111 0100 -> int64
		var bs []byte
		bs, err = readBytes(buf, cursor, 8)
		if err != nil {
			return nil, 0, err
		}
		result = []byte(strconv.FormatInt(int64(binary.LittleEndian.Uint64(bs)), 10))
		length = readVarInt(buf, cursor)
		return result, length, nil
	case 15: // 1111 1111 -> end
		return nil, 0, errors.New("unexpected end")
	}
	return nil, 0, fmt.Errorf("unknown entry header")
}
