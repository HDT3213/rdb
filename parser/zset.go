package parser

import "strconv"

func (p *Parser) readZSet(zset2 bool) ([]*ZSetEntry, error) {
	length, _, err := p.readLength()
	if err != nil {
		return nil, err
	}
	entries := make([]*ZSetEntry, 0, int(length))
	for i := uint64(0); i < length; i++ {
		member, err := p.readString()
		if err != nil {
			return nil, err
		}
		var score float64
		if zset2 {
			score, err = p.readFloat()
		} else {
			score, err = p.readLiteralFloat()
		}
		entries = append(entries, &ZSetEntry{
			Member: string(member),
			Score:  score,
		})
	}
	return entries, nil
}

func (p *Parser) readZipListZSet() ([]*ZSetEntry, error) {
	buf, err := p.readString()
	if err != nil {
		return nil, err
	}
	cursor := 0
	size := readZipListLength(buf, &cursor)
	entries := make([]*ZSetEntry, 0, size)
	for i := 0; i < size; i += 2 {
		member, err := p.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		scoreLiteral, err := p.readZipListEntry(buf, &cursor)
		if err != nil {
			return nil, err
		}
		score, err := strconv.ParseFloat(string(scoreLiteral), 64)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &ZSetEntry{
			Member: string(member),
			Score:  score,
		})
	}
	return entries, nil
}
