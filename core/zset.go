package core

import (
	"github.com/hdt3213/rdb/model"
	"strconv"
)

func (dec *Decoder) readZSet(zset2 bool) ([]*model.ZSetEntry, error) {
	length, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	entries := make([]*model.ZSetEntry, 0, int(length))
	for i := uint64(0); i < length; i++ {
		member, err := dec.readString()
		if err != nil {
			return nil, err
		}
		var score float64
		if zset2 {
			score, err = dec.readFloat()
		} else {
			score, err = dec.readLiteralFloat()
		}
		entries = append(entries, &model.ZSetEntry{
			Member: string(member),
			Score:  score,
		})
	}
	return entries, nil
}

func (dec *Decoder) readZipListZSet() ([]*model.ZSetEntry, error) {
	buf, err := dec.readString()
	if err != nil {
		return nil, err
	}
	cursor := 0
	size := readZipListLength(buf, &cursor)
	entries := make([]*model.ZSetEntry, 0, size)
	for i := 0; i < size; i += 2 {
		member, err := dec.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		scoreLiteral, err := dec.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		score, err := strconv.ParseFloat(string(scoreLiteral), 64)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &model.ZSetEntry{
			Member: string(member),
			Score:  score,
		})
	}
	return entries, nil
}
