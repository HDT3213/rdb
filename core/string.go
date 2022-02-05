package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/hdt3213/rdb/lzf"
	"math"
	"strconv"
)

const (
	len6Bit      = 0
	len14Bit     = 1
	len32or64Bit = 2
	lenSpecial   = 3
	len32Bit     = 0x80
	len64Bit     = 0x81

	encodeInt8  = 0
	encodeInt16 = 1
	encodeInt32 = 2
	encodeLZF   = 3
)

// readLength parse Length Encoding
// see: https://github.com/sripathikrishnan/redis-rdb-tools/wiki/Redis-RDB-Dump-File-Format#length-encoding
func (dec *Decoder) readLength() (uint64, bool, error) {
	firstByte, err := dec.readByte()
	if err != nil {
		return 0, false, fmt.Errorf("read length failed: %v", err)
	}
	lenType := (firstByte & 0xc0) >> 6 // get first 2 bits
	var length uint64
	special := false
	switch lenType {
	case len6Bit:
		length = uint64(firstByte) & 0x3f
	case len14Bit:
		nextByte, err := dec.readByte()
		if err != nil {
			return 0, false, fmt.Errorf("read len14Bit failed: %v", err)
		}
		length = (uint64(firstByte)&0x3f)<<8 | uint64(nextByte)
	case len32or64Bit:
		if firstByte == len32Bit {
			err = dec.readFull(dec.buffer[0:4])
			if err != nil {
				return 0, false, fmt.Errorf("read len32Bit failed: %v", err)
			}
			length = uint64(binary.BigEndian.Uint32(dec.buffer))
		} else if firstByte == len64Bit {
			err = dec.readFull(dec.buffer)
			if err != nil {
				return 0, false, fmt.Errorf("read len64Bit failed: %v", err)
			}
			length = binary.BigEndian.Uint64(dec.buffer)
		} else {
			return 0, false, fmt.Errorf("illegal length encoding: %x", firstByte)
		}
	case lenSpecial:
		special = true
		length = uint64(firstByte) & 0x3f
	}
	return length, special, nil
}

func (dec *Decoder) readString() ([]byte, error) {
	length, special, err := dec.readLength()
	if err != nil {
		return nil, err
	}

	if special {
		switch length {
		case encodeInt8:
			b, err := dec.readByte()
			return []byte(strconv.Itoa(int(b))), err
		case encodeInt16:
			b, err := dec.readUint16()
			return []byte(strconv.Itoa(int(b))), err
		case encodeInt32:
			b, err := dec.readUint32()
			return []byte(strconv.Itoa(int(b))), err
		case encodeLZF:
			return dec.readLZF()
		default:
			return []byte{}, errors.New("Unknown string encode type ")
		}
	}

	res := make([]byte, length)
	err = dec.readFull(res)
	return res, err
}

func (dec *Decoder) readUint16() (uint16, error) {
	err := dec.readFull(dec.buffer[:2])
	if err != nil {
		return 0, fmt.Errorf("read uint16 error: %v", err)
	}

	i := binary.LittleEndian.Uint16(dec.buffer[:2])
	return i, nil
}

func (dec *Decoder) readUint32() (uint32, error) {
	err := dec.readFull(dec.buffer[:4])
	if err != nil {
		return 0, fmt.Errorf("read uint16 error: %v", err)
	}

	i := binary.LittleEndian.Uint32(dec.buffer[:4])
	return i, nil
}

func (dec *Decoder) readLiteralFloat() (float64, error) {
	first, err := dec.readByte()
	if err != nil {
		return 0, err
	}
	if first == 0xff {
		return math.Inf(-1), nil
	} else if first == 0xfe {
		return math.Inf(1), nil
	} else if first == 0xfd {
		return math.NaN(), nil
	}
	buf := make([]byte, first)
	err = dec.readFull(buf)
	if err != nil {
		return 0, err
	}
	str := string(buf)
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("")
	}
	return val, err
}

func (dec *Decoder) readFloat() (float64, error) {
	err := dec.readFull(dec.buffer)
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(dec.buffer)
	return math.Float64frombits(bits), nil
}

func (dec *Decoder) readLZF() ([]byte, error) {
	inLen, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	outLen, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	val := make([]byte, inLen)
	err = dec.readFull(val)
	if err != nil {
		return nil, err
	}
	result := lzf.Decompress(val, int(inLen), int(outLen))
	return result, nil
}
