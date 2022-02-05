package core

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

func (dec *Decoder) readSet() ([][]byte, error) {
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

func (dec *Decoder) readIntSet() (result [][]byte, err error) {
	var buf []byte
	buf, err = dec.readString()
	if err != nil {
		return nil, err
	}
	sizeBytes := buf[0:4]
	intSize := int(binary.LittleEndian.Uint32(sizeBytes))
	if intSize != 2 && intSize != 4 && intSize != 8 {
		return nil, fmt.Errorf("unknown intset encoding: %d", intSize)
	}
	lenBytes := buf[4:8]
	cardinality := binary.LittleEndian.Uint32(lenBytes)
	cursor := 8
	result = make([][]byte, 0, cardinality)
	for i := uint32(0); i < cardinality; i++ {
		var intBytes []byte
		intBytes, err = readBytes(buf, &cursor, intSize)
		if err != nil {
			return
		}
		var intString string
		switch intSize {
		case 2:
			intString = strconv.FormatInt(int64(int16(binary.LittleEndian.Uint16(intBytes))), 10)
		case 4:
			intString = strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(intBytes))), 10)
		case 8:
			intString = strconv.FormatInt(int64(int64(binary.LittleEndian.Uint64(intBytes))), 10)
		}
		result = append(result, []byte(intString))
	}
	return
}
